// W-TinyLFU eviction backend.
//
// The cache is split into three segments:
//
//   - Window (1 %): a small LRU that absorbs new keys with no admission
//     barrier. When the window overflows its candidate is offered to the main
//     cache.
//   - Probation (≈20 % of main): newly admitted keys from the window land
//     here.
//   - Protected (≈80 % of main): keys that have been accessed while on
//     probation are promoted here.
//
// Admission from window → main is gated by the TinyLFU estimator: the
// incoming candidate's estimated frequency is compared against the probation
// victim's (the LRU end of probation); the one with the lower count is
// evicted.  This keeps frequently-accessed keys resident even if they were
// accessed a long time ago (unlike pure LRU), while still admitting new
// popular keys quickly through the window.
//
// A 4-row count-min sketch with 4-bit saturating counters stores approximate
// access frequencies.  A Bloom-filter "door-keeper" sits in front: a key must
// be seen at least twice before it is counted in the sketch, preventing
// one-hit wonders from polluting frequency estimates.  Both structures are
// reset (halved / cleared) when the total count reaches 10×capacity, keeping
// the estimate fresh ("aging").
//
// Reference:
//
//	"TinyLFU: A Highly Efficient Cache Admission Policy"
//	Gil Einziger, Roy Friedman, Ben Manes (2017)
//	https://dl.acm.org/doi/10.1145/3149371
package size

import (
	"hash/maphash"

	"github.com/upfluence/pkg/v2/cache/policy"
	"github.com/upfluence/pkg/v2/cache/policy/internal/lru"
)

type tnode[K comparable] = lru.Node[K, *lfuSegment[K]]

type lfuSegment[K comparable] struct {
	l   *lru.List[K, *lfuSegment[K]]
	cap int
}

func newLFUSegment[K comparable](cap int) *lfuSegment[K] {
	return &lfuSegment[K]{
		l:   lru.NewList[K, *lfuSegment[K]](),
		cap: cap,
	}
}

func (s *lfuSegment[K]) full() bool             { return s.l.Len >= s.cap }
func (s *lfuSegment[K]) len() int               { return s.l.Len }
func (s *lfuSegment[K]) victim() *tnode[K]      { return s.l.Front() }
func (s *lfuSegment[K]) remove(n *tnode[K])     { s.l.Remove(n) }
func (s *lfuSegment[K]) moveToBack(n *tnode[K]) { s.l.MoveToBack(n) }

func (s *lfuSegment[K]) pushBack(n *tnode[K]) {
	n.Extra = s
	s.l.PushBack(n)
}

// ---------------------------------------------------------------------------
// count-min sketch with 4-bit saturating counters
// ---------------------------------------------------------------------------

type lfuSketch struct {
	rows [4][]uint8
	w    uint64
	mask uint64

	total   int
	resetAt int
}

func newLFUSketch(capacity int) lfuSketch {
	w := uint64(16)
	target := uint64(capacity) * 10
	for w < target {
		w <<= 1
	}
	s := lfuSketch{
		w:       w,
		mask:    w - 1,
		resetAt: capacity * 10,
	}
	for i := range s.rows {
		s.rows[i] = make([]uint8, w/2)
	}
	return s
}

func (s *lfuSketch) increment(h1, h2 uint64) {
	s.total++
	if s.total >= s.resetAt {
		s.halve()
	}

	for i, row := range s.rows {
		idx := (h1 + uint64(i)*h2) & s.mask
		byteIdx := idx >> 1
		if idx&1 == 0 {
			if row[byteIdx]&0x0f < 0x0f {
				row[byteIdx]++
			}
		} else {
			if row[byteIdx]&0xf0 < 0xf0 {
				row[byteIdx] += 0x10
			}
		}
	}
}

func (s *lfuSketch) estimate(h1, h2 uint64) uint8 {
	var min uint8 = 0xff
	for i, row := range s.rows {
		idx := (h1 + uint64(i)*h2) & s.mask
		byteIdx := idx >> 1
		var v uint8
		if idx&1 == 0 {
			v = row[byteIdx] & 0x0f
		} else {
			v = (row[byteIdx] >> 4) & 0x0f
		}
		if v < min {
			min = v
		}
	}
	return min
}

func (s *lfuSketch) halve() {
	for _, row := range s.rows {
		for i, b := range row {
			row[i] = ((b >> 1) & 0x07) | ((b >> 1) & 0x70)
		}
	}
	s.total >>= 1
}

// ---------------------------------------------------------------------------
// door-keeper: a Bloom filter used as a one-time admission gate.
// ---------------------------------------------------------------------------

type lfuDoorkeeper struct {
	bits    []uint64
	m       uint64
	mask    uint64
	resetAt int
	total   int
}

func newLFUDoorkeeper(capacity int) lfuDoorkeeper {
	m := uint64(16)
	target := uint64(capacity) * 10
	for m < target {
		m <<= 1
	}
	return lfuDoorkeeper{
		bits:    make([]uint64, m/64),
		m:       m,
		mask:    m - 1,
		resetAt: capacity * 10,
	}
}

func (d *lfuDoorkeeper) contains(h1, h2 uint64) bool {
	d.total++
	if d.total >= d.resetAt {
		d.reset()
	}

	pos1 := h1 & d.mask
	pos2 := h2 & d.mask

	w1, b1 := pos1/64, pos1%64
	w2, b2 := pos2/64, pos2%64

	seen := (d.bits[w1]>>b1)&1 == 1 && (d.bits[w2]>>b2)&1 == 1
	d.bits[w1] |= 1 << b1
	d.bits[w2] |= 1 << b2
	return seen
}

func (d *lfuDoorkeeper) reset() {
	for i := range d.bits {
		d.bits[i] = 0
	}
	d.total = 0
}

type tinylfuBackend[K comparable] struct {
	window    *lfuSegment[K]
	protected *lfuSegment[K]
	probation *lfuSegment[K]

	sketch     lfuSketch
	doorkeeper lfuDoorkeeper

	seed1, seed2 maphash.Seed
}

func (b *tinylfuBackend[K]) hash(k K) (uint64, uint64) {
	h1 := maphash.Comparable(b.seed1, k)
	h2 := maphash.Comparable(b.seed2, k)
	h2 |= 1
	return h1, h2
}

func (b *tinylfuBackend[K]) incrementFreq(h1, h2 uint64) {
	if b.doorkeeper.contains(h1, h2) {
		b.sketch.increment(h1, h2)
	}
}

// Insert adds k to the window. If the window overflows the LRU window
// candidate is offered to main via the TinyLFU admission test.
// The returned evicted key (ok=true) is whichever key lost the contest.
func (b *tinylfuBackend[K]) insert(k K) (K, bool, *tnode[K]) {
	h1, h2 := b.hash(k)
	b.incrementFreq(h1, h2)

	n := b.window.l.Alloc()
	n.Key = k
	b.window.pushBack(n)

	if !b.window.full() {
		var zero K
		return zero, false, n
	}

	candidate := b.window.victim()
	b.window.remove(candidate)

	evicted, ok := b.admitToMain(candidate)
	return evicted, ok, n
}

func (b *tinylfuBackend[K]) admitToMain(candidate *tnode[K]) (K, bool) {
	totalMain := b.protected.len() + b.probation.len()
	mainCap := b.protected.cap + b.probation.cap

	if totalMain < mainCap {
		b.probation.pushBack(candidate)
		var zero K
		return zero, false
	}

	victim := b.probation.victim()

	ch1, ch2 := b.hash(candidate.Key)
	vh1, vh2 := b.hash(victim.Key)

	if b.sketch.estimate(ch1, ch2) > b.sketch.estimate(vh1, vh2) {
		evicted := victim.Key
		b.probation.remove(victim)
		b.window.l.Free(victim)
		b.probation.pushBack(candidate)
		return evicted, true
	}

	evicted := candidate.Key
	b.window.l.Free(candidate)
	return evicted, true
}

// remove deletes n from its segment (explicit Evict op).
func (b *tinylfuBackend[K]) remove(n *tnode[K]) {
	n.Extra.remove(n)
	b.window.l.Free(n)
}

// get records a hit: increments frequency and promotes n.
func (b *tinylfuBackend[K]) get(k K, n *tnode[K]) {
	h1, h2 := b.hash(k)
	b.incrementFreq(h1, h2)
	b.promote(n)
}

func (b *tinylfuBackend[K]) promote(n *tnode[K]) {
	switch n.Extra {
	case b.window, b.protected:
		n.Extra.moveToBack(n)

	case b.probation:
		b.probation.remove(n)

		if b.protected.full() {
			demoted := b.protected.victim()
			b.protected.remove(demoted)
			b.probation.pushBack(demoted)
		}

		b.protected.pushBack(n)
	}
}

// NewTinyLFUPolicy returns a W-TinyLFU eviction policy for a cache of the
// given capacity (must be ≥ 1).
//
// Segment sizing follows Caffeine's defaults:
//
//	window    =  1 % of capacity (min 1)
//	protected = 80 % of main    (min 1)
//	probation = 20 % of main    (min 1)
func NewTinyLFUPolicy[K comparable](capacity int) policy.EvictionPolicy[K] {
	if capacity < 1 {
		capacity = 1
	}

	windowCap := capacity / 100
	if windowCap < 1 {
		windowCap = 1
	}

	mainCap := capacity - windowCap
	if mainCap < 2 {
		mainCap = 2
	}

	protectedCap := mainCap * 8 / 10
	if protectedCap < 1 {
		protectedCap = 1
	}

	probationCap := mainCap - protectedCap
	if probationCap < 1 {
		probationCap = 1
	}

	return newPolicy[K, *tnode[K]](
		&tinylfuBackend[K]{
			window:     newLFUSegment[K](windowCap),
			protected:  newLFUSegment[K](protectedCap),
			probation:  newLFUSegment[K](probationCap),
			sketch:     newLFUSketch(capacity),
			doorkeeper: newLFUDoorkeeper(capacity),
			seed1:      maphash.MakeSeed(),
			seed2:      maphash.MakeSeed(),
		},
		capacity,
	)
}

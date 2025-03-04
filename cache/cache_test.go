package cache

import (
	"fmt"
	"hash/maphash"
	"math/rand/v2"
	"testing"
)

const (
	offset64 uint64 = 14695981039346656037
	prime64  uint64 = 1099511628211
)

func generateTestStrings(size, count int) []string {
	result := make([]string, count)

	for i := 0; i < count; i++ {
		bytes := make([]byte, size)
		for j := range bytes {
			bytes[j] = byte(rand.IntN(256))
		}

		result[i] = string(bytes)
	}

	return result
}

func fnv64aHash(s string) uint64 {
	h := offset64

	for _, b := range []byte(s) {
		h ^= uint64(b)
		h *= prime64
	}

	return h
}

func BenchmarkHashFnv64(b *testing.B) {
	sizes := []int{8, 16, 64, 256}
	for _, size := range sizes {
		testData := generateTestStrings(size, 1000)

		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				for _, s := range testData {
					fnv64aHash(s)
				}
			}
		})
	}
}

func BenchmarkHashMaphash(b *testing.B) {
	sizes := []int{8, 16, 64, 256}
	for _, size := range sizes {
		testData := generateTestStrings(size, 1000)
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			b.ResetTimer()
			seed := maphash.MakeSeed()

			for range b.N {
				for _, s := range testData {
					maphash.String(seed, s)
				}
			}
		})
	}
}

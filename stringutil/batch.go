package stringutil

// Deprecated: Use generics.Batch instead.
func Batch(slice []string, size int) [][]string {
	if len(slice) == 0 {
		return nil
	}

	if size < 1 {
		panic("stringutil: illegal batch size")
	}

	batches := make([][]string, 0, (len(slice)+size-1)/size)

	for size < len(slice) {
		slice, batches = slice[size:], append(batches, slice[0:size:size])
	}

	batches = append(batches, slice)

	return batches
}

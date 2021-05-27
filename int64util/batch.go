package int64util

func Batch(slice []int64, size int) [][]int64 {
	if len(slice) == 0 {
		return nil
	}

	if size < 1 {
		panic("int64util: illegal batch size")
	}

	batches := make([][]int64, 0, (len(slice)+size-1)/size)

	for size < len(slice) {
		slice, batches = slice[size:], append(batches, slice[0:size:size])
	}

	batches = append(batches, slice)

	return batches
}

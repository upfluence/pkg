package slices

func Batch[T any](slice []T, size int) [][]T {
	if len(slice) == 0 {
		return nil
	}

	if size < 1 {
		panic("generics: illegal batch size")
	}

	batches := make([][]T, 0, (len(slice)+size-1)/size)

	for size < len(slice) {
		batches = append(batches, slice[:size:size])
		slice = slice[size:]
	}

	batches = append(batches, slice)

	return batches
}

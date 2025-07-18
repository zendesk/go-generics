package functions

// Batch splits a slice into smaller slices of a specified size. If the batchSize <= 0, it panics
func Batch[T any](items []T, batchSize int) [][]T {
	if batchSize <= 0 {
		panic("Invalid batch size of <= 0.")
	}

	if len(items) == 0 {
		return nil
	}

	var batches [][]T
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}
	return batches
}

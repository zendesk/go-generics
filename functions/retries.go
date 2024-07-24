package functions

import "time"

// RunWithRetries will run fn numRetries+1 times, if an error is returned, the fn will be run again, until
// numRetries is exceeded or a non nil error is returned.
func RunWithRetries[T any](fn func(t T) error, item T, numRetries int, backoffInterval time.Duration) error {
	var err error
	for i := 0; i < numRetries+1; i++ {
		err = fn(item)

		if err == nil {
			break
		}

		// Sleep for backoffInterval * number of backoffs
		time.Sleep(backoffInterval * time.Duration(i+1))
	}

	return err
}

// runMapWithRetries will run fn numRetries+1 times, if an error is returned, the fn will be run again, until
// numRetries is exceeded or a non-nil error is returned
func runMapWithRetries[T any, Y any](fn func(T) (Y, error), item T, numRetries int, backoffInterval time.Duration) (Y, error) {
	var err error
	var result Y
	for i := 0; i < numRetries+1; i++ {
		result, err = fn(item)

		if err == nil {
			break
		}

		time.Sleep(backoffInterval * time.Duration(i+1))
	}

	return result, err
}

// runToMapWithRetries will run fn numRetries+1 times, if an error is returned, the fn will be run again, until
// numRetries is exceeded or a non-nil error is returned
func runToMapWithRetries[T any, K comparable, V any](fn func(T) (K, V, error), item T, numRetries int, backoffInterval time.Duration) (K, V, error) {
	var err error
	var k K
	var v V
	for i := 0; i < numRetries+1; i++ {
		k, v, err = fn(item)

		if err == nil {
			break
		}

		time.Sleep(backoffInterval * time.Duration(i+1))
	}

	return k, v, err
}

// runMapToManyWithRetries will run fn numRetries+1 times, if an error is returned, the fn will be run again, until
// numRetries is exceeded or a non-nil error is returned
func runMapToManyWithRetries[T any, Y any](fn func(T) ([]Y, error), item T, numRetries int, backoffInterval time.Duration) ([]Y, error) {
	var err error
	var result []Y
	for i := 0; i < numRetries+1; i++ {
		result, err = fn(item)

		if err == nil {
			break
		}

		time.Sleep(backoffInterval * time.Duration(i+1))
	}

	return result, err
}

// runMapToSliceWithRetries will run fn numRetries+1 times, if an error is returned, the fn will be run again, until
// numRetries is exceeded or a non-nil error is returned
func runMapToSliceWithRetries[K comparable, V any, Z any](fn func(K, V) (Z, error), k K, v V, numRetries int, backoffInterval time.Duration) (Z, error) {
	var err error
	var result Z
	for i := 0; i < numRetries+1; i++ {
		result, err = fn(k, v)

		if err == nil {
			break
		}

		time.Sleep(backoffInterval * time.Duration(i+1))
	}

	return result, err
}

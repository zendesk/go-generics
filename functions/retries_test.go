package functions

import (
	"fmt"
	"testing"
	"time"

	"github.com/zendesk/go-generics/functions/internal/test"
)

func TestRunEachWithRetries(t *testing.T) {
	maxRetries := 4
	backoffInterval := time.Millisecond * 10

	// Total backoff expected = 10 + 20 + 30 + 40
	minDuration := time.Millisecond * 100
	count := 0
	start := time.Now()
	testFunc := func(i int) error {
		if count < maxRetries {
			count++
			return fmt.Errorf("i is %d, this is an error to force retry!", i)
		}
		return nil
	}

	err := RunWithRetries(testFunc, 0, maxRetries, backoffInterval)
	finish := time.Now()
	test.CheckErr(err, "Unexpected error with RunEachWithRetries test.", t)

	if finish.Sub(start) < minDuration {
		t.Fatalf("Retries did not backoff a minimum expected duration.")
	}
}

func TestRunMapWithRetries(t *testing.T) {
	maxRetries := 4
	backoffInterval := time.Millisecond * 10

	// Total backoff expected = 10 + 20 + 30 + 40
	minDuration := time.Millisecond * 100
	input := 1
	count := 0
	expectedResult := maxRetries * input
	start := time.Now()

	testFunc := func(i int) (int, error) {
		if count < maxRetries {
			count++
			return 0, fmt.Errorf("i is %d, this is an error to force retry", i)
		}
		return count * i, nil
	}

	result, err := runMapWithRetries(testFunc, input, maxRetries, backoffInterval)
	finish := time.Now()
	test.CheckErr(err, "Unexpected error with runMapWithRetries test.", t)

	if finish.Sub(start) < minDuration {
		t.Fatalf("Retries did not backoff a minimum expected duration.")
	}

	test.CheckEqual(result, "Total Result", expectedResult, t)
}
func TestRunToMapWithRetries(t *testing.T) {
	maxRetries := 4
	backoffInterval := time.Millisecond * 10

	// Total backoff expected = 10 + 20 + 30 + 40
	minDuration := time.Millisecond * 100
	input := 1
	count := 0
	expectedResult := maxRetries * input
	start := time.Now()

	testFunc := func(i int) (int, int, error) {
		if count < maxRetries {
			count++
			return 0, 0, fmt.Errorf("i is %d, this is an error to force retry", i)
		}
		return count * i, count * i * 2, nil
	}

	key, val, err := runToMapWithRetries(testFunc, input, maxRetries, backoffInterval)
	finish := time.Now()
	test.CheckErr(err, "Unexpected error with runMapWithRetries test.", t)

	if finish.Sub(start) < minDuration {
		t.Fatalf("Retries did not backoff a minimum expected duration.")
	}

	test.CheckEqual(key, "Key", expectedResult, t)
	test.CheckEqual(val, "Val", expectedResult*2, t)
}

func TestRunMapToManyWithRetries(t *testing.T) {
	maxRetries := 4
	backoffInterval := time.Millisecond * 10

	// Total backoff expected = 10 + 20 + 30 + 40
	minDuration := time.Millisecond * 100
	input := map[int]int{
		0: 1,
	}
	count := 0
	expectedResult := []int{1}
	start := time.Now()

	testFunc := func(i map[int]int) ([]int, error) {
		if count < maxRetries {
			count++
			return []int{9999}, fmt.Errorf("i is %d, this is an error to force retry", i)
		}
		return []int{i[0]}, nil
	}

	results, err := runMapToManyWithRetries(testFunc, input, maxRetries, backoffInterval)
	finish := time.Now()
	test.CheckErr(err, "Unexpected error with runMapWithRetries test.", t)

	if finish.Sub(start) < minDuration {
		t.Fatalf("Retries did not backoff a minimum expected duration.")
	}

	test.CheckEqual(results, "results", expectedResult, t)
}

func TestRunMapToSliceWithRetries(t *testing.T) {
	maxRetries := 4
	backoffInterval := time.Millisecond * 10

	// Total backoff expected = 10 + 20 + 30 + 40
	minDuration := time.Millisecond * 100
	key := 0
	val := 1
	count := 0
	expectedResult := []int{val}
	start := time.Now()

	testFunc := func(k, v int) ([]int, error) {
		if count < maxRetries {
			count++
			return []int{9999}, fmt.Errorf("k is %d, this is an error to force retry", k)
		}
		return []int{v}, nil
	}

	results, err := runMapToSliceWithRetries(testFunc, key, val, maxRetries, backoffInterval)
	finish := time.Now()
	test.CheckErr(err, "Unexpected error with runMapWithRetries test.", t)

	if finish.Sub(start) < minDuration {
		t.Fatalf("Retries did not backoff a minimum expected duration.")
	}

	test.CheckEqual(results, "results", expectedResult, t)
}

func TestGoMapEachWithRetries(t *testing.T) {
	maxRetries := 4
	backoffInterval := time.Millisecond * 10

	// Total backoff expected = 10 + 20 + 30 + 40
	minDuration := time.Millisecond * 100 // 100 total b/c of concurrency
	input := 1
	count := 0
	items := []int{1}
	expectedResult := maxRetries * input * len(items)
	start := time.Now()

	testFunc := func(i int) (int, error) {
		if count < maxRetries {
			count++
			return 0, fmt.Errorf("i is %d, this is an error to force retry", i)
		}

		resultVal := count * i
		count = 0
		return resultVal, nil
	}

	results, errs := GoMapWithErrs(items, testFunc, RetryOption(maxRetries, backoffInterval))
	finish := time.Now()

	if len(errs) > 0 {
		t.Fatalf("Unexpected errors found: %+v", errs)
	}

	if finish.Sub(start) < minDuration {
		t.Fatalf("Retries did not backoff a minimum expected duration.")
	}

	test.CheckEqual(results, "Total Result", []int{expectedResult}, t)
}

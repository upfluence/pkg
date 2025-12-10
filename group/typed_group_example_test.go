package group_test

import (
	"context"
	"fmt"

	"github.com/upfluence/pkg/v2/group"
)

func ExampleTypedGroup() {
	ctx := context.Background()

	tg := &group.TypedGroup[int]{
		Group: group.WaitGroup(ctx),
		Value: 0,
	}

	// Concurrently sum values
	for i := 1; i <= 10; i++ {
		tg.Do(func(ctx context.Context) (func(*int), error) {
			return func(sum *int) {
				// We don't need to lock here because TypedGroup do it for us
				*sum += i
			}, nil
		})
	}

	total, err := tg.Wait()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("total: %d\n", total)
	// Output: total: 55
}

func ExampleTypedGroup_second() {
	ctx := context.Background()

	tg := &group.TypedGroup[map[string]int]{
		Group: group.WaitGroup(ctx),
		Value: make(map[string]int),
	}

	// Simulate counting items from different sources
	items := []int{10, 20, 30}

	for _, count := range items {
		tg.Do(func(ctx context.Context) (func(*map[string]int), error) {
			// Simulate countItems
			return func(m *map[string]int) {
				(*m)["total"] += count
			}, nil
		})
	}

	result, err := tg.Wait()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("total count: %d\n", result["total"])
	// Output: total count: 60
}

func ExampleTypedRunner() {
	ctx := context.Background()

	tg := &group.TypedGroup[[]string]{
		Group: group.WaitGroup(ctx),
		Value: []string{},
	}

	// Simulate fetching data from different sources
	dataSources := [][]string{
		{"apple", "banana"},
		{"cherry", "date"},
		{"elderberry"},
	}

	for _, data := range dataSources {
		var runner group.TypedRunner[[]string] = func(ctx context.Context) (func(*[]string), error) {
			// Simulate fetchData
			fetchedData := data
			// Return a function to merge data into the accumulator
			return func(acc *[]string) {
				*acc = append(*acc, fetchedData...)
			}, nil
		}
		tg.Do(runner)
	}

	result, err := tg.Wait()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("collected %d items\n", len(result))
	// Output: collected 5 items
}

package group_test

import (
	"context"
	"fmt"

	"github.com/upfluence/errors"

	"github.com/upfluence/pkg/v2/group"
)

func ExampleWaitGroup() {
	var err = errors.New("sample error")

	g := group.WaitGroup(context.Background())

	// Validate multiple items, collect all validation errors
	for i := range 4 {
		g.Do(func(ctx context.Context) error {
			if i%2 == 0 {
				return err
			}

			return nil
		})
	}

	// Returns all errors encountered
	if err := g.Wait(); err != nil {
		fmt.Printf("validation failed: %v", err)
	}
	// Output: validation failed: [sample error, sample error]
}

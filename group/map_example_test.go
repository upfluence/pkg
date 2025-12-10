package group

import (
	"context"
	"fmt"
	"strconv"
)

func ExampleExecuteMap() {
	res, err := ExecuteMap[int, string](
		WaitGroup(context.Background()),
		[]int{1, 2, 3},
		func(_ context.Context, v int) (string, error) {
			return "id-" + strconv.Itoa(v), nil
		},
	)

	if err != nil {
		panic(err)
	}

	fmt.Println(res)
	// Output: map[1:id-1 2:id-2 3:id-3]
}

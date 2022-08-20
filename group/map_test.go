package group

import (
	"context"
	"reflect"
	"strconv"
	"testing"
)

func TestExecuteMap(t *testing.T) {
	res, err := ExecuteMap[int, string](
		ErrorGroup(context.Background()),
		[]int{1, 2, 3},
		func(_ context.Context, v int) (string, error) {
			return strconv.Itoa(v), nil
		},
	)

	if err != nil {
		t.Errorf("ExecuteMap(...) = (_, %v), wanted nil", err)
	}

	if !reflect.DeepEqual(res, map[int]string{1: "1", 2: "2", 3: "3"}) {
		t.Errorf("ExecuteMap(...) = (%v, nil), wanted [1, 2, 3]", res)
	}
}

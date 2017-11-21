package metricutil

import "fmt"

const successStatus = "success"

var errorStatusMap = make(map[error]string)

func UpsertErrorStatus(err error, status string) {
	errorStatusMap[err] = status
}

func ResultStatus(err error) string {
	if err == nil {
		return successStatus
	}

	if status, ok := errorStatusMap[err]; ok {
		return status
	}

	if t := fmt.Sprintf("%T", err); t != "*errors.errorString" {
		return t
	}

	return err.Error()
}

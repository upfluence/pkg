package cfg

import (
	"os"
	"strconv"

	"github.com/upfluence/goutils/log"
)

func FetchString(variable, defaultValue string) string {
	if v := os.Getenv(variable); v != "" {
		return v
	}

	return defaultValue
}

func FetchInt(variable string, defaultValue int) int {
	if v := os.Getenv(variable); v != "" {
		v1, err := strconv.Atoi(v)

		if err == nil {
			return v1
		}

		log.Errorf("fetchInt: %s", err.Error())
	}

	return defaultValue
}

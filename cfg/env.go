package cfg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const listSeparator = ","

func FetchBool(variable string, defaultValue bool) bool {
	envV := os.Getenv(variable)

	for _, v := range []string{"true", "t", "1"} {
		if v == envV {
			return true
		}
	}

	for _, v := range []string{"false", "f", "0"} {
		if v == envV {
			return false
		}
	}

	return defaultValue
}

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

		fmt.Printf("cfg: fetchInt: %s\n", err.Error())
	}

	return defaultValue
}

func FetchStrings(variable string, defaultValue []string) []string {
	if v := os.Getenv(variable); v != "" {
		return strings.Split(v, listSeparator)
	}

	return defaultValue
}

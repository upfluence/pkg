package testutil

import (
	"os"
	"testing"
)

func FetchEnvVariable(t testing.TB, ev string) string {
	v := os.Getenv(ev)

	if v == "" {
		t.Skipf("%q env var is not defined, skipping test", ev)
	}

	return v
}

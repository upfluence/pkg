package redis

import (
	"os"
	"testing"

	"github.com/upfluence/pkg/stringcache/testutil"
)

func TestIntegration(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		t.Skip("no REDIS_URL provided")
	}

	c, err := NewCache(redisURL, "test-key")

	if err != nil {
		t.Errorf("Error returned by the constructor %+v", err)
	}

	c.conn.Do("SDEL", c.key)

	testutil.IntegrationScenario(t, c)
}

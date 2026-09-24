package cloudwatchlog

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlyStreamName(t *testing.T) {
	t.Run("uses FLY_MACHINE_ID when set", func(t *testing.T) {
		t.Setenv("FLY_MACHINE_ID", "e784edb3a91d89")

		assert.Equal(t, "e784edb3a91d89", FlyStreamName())
	})

	t.Run("falls back to hostname+ulid when unset", func(t *testing.T) {
		t.Setenv("FLY_MACHINE_ID", "")

		host, err := os.Hostname()
		if err != nil || host == "" {
			host = "unknown"
		}

		got := FlyStreamName()
		assert.True(t, strings.HasPrefix(got, host+"-"), "expected %q to start with %q", got, host+"-")

		got2 := FlyStreamName()
		assert.NotEqual(t, got, got2, "expected successive calls to produce unique stream names")
	})
}

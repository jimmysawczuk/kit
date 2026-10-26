package webidentity

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitForToken(t *testing.T) {
	t.Run("returns immediately when the variable is unset", func(t *testing.T) {
		t.Setenv(TokenFileEnv, "")

		waited, ok := WaitForToken(context.Background(), time.Minute)
		assert.True(t, ok)
		assert.Zero(t, waited)
	})

	t.Run("returns immediately when the file exists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("jwt"), 0o600))
		t.Setenv(TokenFileEnv, path)

		waited, ok := WaitForToken(context.Background(), time.Minute)
		assert.True(t, ok)
		assert.Zero(t, waited)
	})

	t.Run("waits for the file to appear", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		t.Setenv(TokenFileEnv, path)

		go func() {
			time.Sleep(250 * time.Millisecond)
			_ = os.WriteFile(path, []byte("jwt"), 0o600)
		}()

		waited, ok := WaitForToken(context.Background(), 5*time.Second)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, waited, 250*time.Millisecond)
		assert.Less(t, waited, 5*time.Second)
	})

	t.Run("gives up after timeout", func(t *testing.T) {
		t.Setenv(TokenFileEnv, filepath.Join(t.TempDir(), "token"))

		waited, ok := WaitForToken(context.Background(), 300*time.Millisecond)
		assert.False(t, ok)
		assert.GreaterOrEqual(t, waited, 300*time.Millisecond)
	})

	t.Run("stops when ctx is done", func(t *testing.T) {
		t.Setenv(TokenFileEnv, filepath.Join(t.TempDir(), "token"))

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		waited, ok := WaitForToken(ctx, time.Minute)
		assert.False(t, ok)
		assert.Less(t, waited, time.Minute)
	})
}

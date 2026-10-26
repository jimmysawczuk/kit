package cloudwatchlog

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/jimmysawczuk/kit/aws/webidentity"
	"github.com/lzap/cloudwatchwriter2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDescribe = errors.New("describe failed")

// failingClient fails the first call NewWriter makes, so tests can observe
// how long NewWriter blocked first and how it wrapped the error.
type failingClient struct {
	cloudwatchwriter2.CloudWatchLogsClient
}

func (failingClient) DescribeLogStreams(context.Context, *cloudwatchlogs.DescribeLogStreamsInput, ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	return nil, errDescribe
}

func TestNewWriter(t *testing.T) {
	t.Run("waits for the token by default and notes it in the error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		t.Setenv(webidentity.TokenFileEnv, path)

		start := time.Now()
		_, err := NewWriter(context.Background(), failingClient{}, "group", "stream", WithTokenWait(300*time.Millisecond))
		require.ErrorIs(t, err, errDescribe)
		assert.GreaterOrEqual(t, time.Since(start), 300*time.Millisecond)
		assert.Contains(t, err.Error(), path)
		assert.Contains(t, err.Error(), "not ready after waiting")
	})

	t.Run("reports the actual wait when ctx ends early", func(t *testing.T) {
		t.Setenv(webidentity.TokenFileEnv, filepath.Join(t.TempDir(), "token"))

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		_, err := NewWriter(ctx, failingClient{}, "group", "stream")
		require.ErrorIs(t, err, errDescribe)
		assert.NotContains(t, err.Error(), DefaultTokenWait.String())
	})

	t.Run("doesn't mention the token once it's present", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("jwt"), 0o600))
		t.Setenv(webidentity.TokenFileEnv, path)

		_, err := NewWriter(context.Background(), failingClient{}, "group", "stream")
		require.ErrorIs(t, err, errDescribe)
		assert.NotContains(t, err.Error(), "token")
	})

	t.Run("WithTokenWait(0) skips the wait", func(t *testing.T) {
		t.Setenv(webidentity.TokenFileEnv, filepath.Join(t.TempDir(), "token"))

		start := time.Now()
		_, err := NewWriter(context.Background(), failingClient{}, "group", "stream", WithTokenWait(0))
		require.ErrorIs(t, err, errDescribe)
		assert.Less(t, time.Since(start), 100*time.Millisecond)
		assert.NotContains(t, err.Error(), "token")
	})

	t.Run("rejects a too-small batch interval before waiting", func(t *testing.T) {
		t.Setenv(webidentity.TokenFileEnv, filepath.Join(t.TempDir(), "token"))

		start := time.Now()
		_, err := NewWriter(context.Background(), failingClient{}, "group", "stream", WithBatchInterval(time.Millisecond))
		require.ErrorIs(t, err, cloudwatchwriter2.ErrBatchIntervalTooSmall)
		assert.Less(t, time.Since(start), 100*time.Millisecond)
	})
}

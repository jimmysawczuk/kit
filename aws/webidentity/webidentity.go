// Package webidentity helps processes that get AWS credentials through a web
// identity token file (AWS_WEB_IDENTITY_TOKEN_FILE), such as apps on Fly.io
// using OIDC, avoid racing the file's creation at startup.
package webidentity

import (
	"context"
	"os"
	"time"
)

// TokenFileEnv is the environment variable the AWS SDK reads the web
// identity token file path from.
const TokenFileEnv = "AWS_WEB_IDENTITY_TOKEN_FILE"

const pollInterval = 100 * time.Millisecond

// WaitForToken blocks until the file named by AWS_WEB_IDENTITY_TOKEN_FILE
// exists, timeout elapses, or ctx is done. Fly's init is supposed to write
// /.fly/oidc_token before starting the app, but some init versions race it,
// and an AWS call made before the file exists fails credential resolution.
//
// Call it before config.LoadDefaultConfig, or at least before the first AWS
// call. It returns immediately if the variable is unset or the file already
// exists.
//
// It returns how long it waited (zero if it didn't have to) and whether the
// file is present, which is also true when the variable is unset.
func WaitForToken(ctx context.Context, timeout time.Duration) (time.Duration, bool) {
	path := os.Getenv(TokenFileEnv)
	if path == "" {
		return 0, true
	}

	return waitForFile(ctx, path, timeout)
}

func waitForFile(ctx context.Context, path string, timeout time.Duration) (time.Duration, bool) {
	if exists(path) {
		return 0, true
	}

	start := time.Now()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(pollInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if exists(path) {
				return time.Since(start), true
			}
		case <-deadline.C:
			return time.Since(start), exists(path)
		case <-ctx.Done():
			return time.Since(start), exists(path)
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

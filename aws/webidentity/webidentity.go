// Package webidentity helps processes that get AWS credentials through a web
// identity token file (AWS_WEB_IDENTITY_TOKEN_FILE), such as apps on Fly.io
// using OIDC, avoid racing the file's creation at startup.
package webidentity

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"time"
)

// TokenFileEnv is the environment variable the AWS SDK reads the web
// identity token file path from.
const TokenFileEnv = "AWS_WEB_IDENTITY_TOKEN_FILE"

const pollInterval = 100 * time.Millisecond

// WaitForToken blocks until the file named by AWS_WEB_IDENTITY_TOKEN_FILE
// has content, timeout elapses, or ctx is done. Fly's init is supposed to write
// /.fly/oidc_token before starting the app, but some init versions race it,
// and an AWS call made before the file exists fails credential resolution.
//
// Call it before config.LoadDefaultConfig, or at least before the first AWS
// call. It returns immediately if the variable is unset or the file already
// has content. The file only counts once it's non-empty, in case it's
// created before the token is written to it. A stat error other than "does
// not exist", such as a permission error, also returns immediately, since
// waiting won't fix it.
//
// It returns how long it waited (zero if it didn't have to) and whether the
// token is present, which is also true when the variable is unset.
func WaitForToken(ctx context.Context, timeout time.Duration) (time.Duration, bool) {
	path := os.Getenv(TokenFileEnv)
	if path == "" {
		return 0, true
	}

	if ready, retry := check(path); ready || !retry {
		return 0, ready
	}

	start := time.Now()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(pollInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			if ready, retry := check(path); ready || !retry {
				return time.Since(start), ready
			}
		case <-deadline.C:
			ready, _ := check(path)
			return time.Since(start), ready
		case <-ctx.Done():
			ready, _ := check(path)
			return time.Since(start), ready
		}
	}
}

// check reports whether path holds a token, and if not, whether it's worth
// polling again.
func check(path string) (ready, retry bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, errors.Is(err, fs.ErrNotExist)
	}

	return fi.Size() > 0, true
}

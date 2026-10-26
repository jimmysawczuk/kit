// Package cloudwatchlog builds an io.WriteCloser that ships log lines to
// AWS CloudWatch Logs, suitable for wiring into zerolog (or anything else
// that writes lines to an io.Writer) via zerolog.MultiLevelWriter.
package cloudwatchlog

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jimmysawczuk/kit/aws/webidentity"
	"github.com/lzap/cloudwatchwriter2"
)

// DefaultBatchInterval is how often queued log events are flushed to
// CloudWatch when no WithBatchInterval option is given.
const DefaultBatchInterval = 5 * time.Second

// DefaultTokenWait is how long NewWriter waits for the web identity token
// file when no WithTokenWait option is given.
const DefaultTokenWait = 5 * time.Second

type settings struct {
	batchInterval time.Duration
	tokenWait     time.Duration
}

// Option configures NewWriter.
type Option func(*settings)

// WithBatchInterval overrides DefaultBatchInterval. It must be at least
// cloudwatchwriter2.MinBatchInterval (200ms); NewWriter returns an error
// otherwise.
func WithBatchInterval(d time.Duration) Option {
	return func(s *settings) { s.batchInterval = d }
}

// WithTokenWait overrides DefaultTokenWait. Zero disables the wait.
func WithTokenWait(d time.Duration) Option {
	return func(s *settings) { s.tokenWait = d }
}

// NewWriter builds an io.WriteCloser that ships log lines to logGroup under
// logStreamName using client, typically a *cloudwatchlogs.Client built from
// the caller's shared aws.Config. The log group and stream are created if
// they're missing, which requires logs:CreateLogGroup and
// logs:CreateLogStream permissions.
//
// ctx's values are passed through, but its cancellation is not: the
// writer's background flushing runs until Close is called.
//
// Callers on Fly can name the stream with FlyStreamName; other callers
// should pass a name that's unique enough to avoid two processes writing to
// the same stream and stomping on each other's sequence tokens.
//
// NewWriter makes an AWS call right away, so if AWS_WEB_IDENTITY_TOKEN_FILE
// is set it first waits up to DefaultTokenWait (see WithTokenWait) for that
// file to exist; see webidentity.WaitForToken. This relies on the SDK's web
// identity credentials provider reading the token file on first use rather
// than when client's config was loaded, which is true as of aws-sdk-go-v2
// credentials v1.20. Callers that want to be independent of that should call
// webidentity.WaitForToken before config.LoadDefaultConfig.
func NewWriter(ctx context.Context, client cloudwatchwriter2.CloudWatchLogsClient, logGroup, streamName string, opts ...Option) (*cloudwatchwriter2.CloudWatchWriter, error) {
	s := settings{batchInterval: DefaultBatchInterval, tokenWait: DefaultTokenWait}
	for _, opt := range opts {
		opt(&s)
	}

	tokenOK := true
	if s.tokenWait > 0 {
		_, tokenOK = webidentity.WaitForToken(ctx, s.tokenWait)
	}

	// cloudwatchwriter2 uses this ctx for its background flush goroutine and
	// every PutLogEvents call, so a cancelled ctx would silently stop shipping.
	writer, err := cloudwatchwriter2.NewWithClientContext(context.WithoutCancel(ctx), client, s.batchInterval, logGroup, streamName)
	if err != nil {
		if !tokenOK {
			return nil, fmt.Errorf("new cloudwatch writer (web identity token file %s missing after %s): %w", os.Getenv(webidentity.TokenFileEnv), s.tokenWait, err)
		}
		return nil, fmt.Errorf("new cloudwatch writer: %w", err)
	}

	return writer, nil
}

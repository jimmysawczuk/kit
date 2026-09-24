// Package cloudwatchlog builds an io.WriteCloser that ships log lines to
// AWS CloudWatch Logs, suitable for wiring into zerolog (or anything else
// that writes lines to an io.Writer) via zerolog.MultiLevelWriter.
package cloudwatchlog

import (
	"context"
	"fmt"
	"time"

	"github.com/lzap/cloudwatchwriter2"
)

// DefaultBatchInterval is how often queued log events are flushed to
// CloudWatch when no WithBatchInterval option is given.
const DefaultBatchInterval = 5 * time.Second

type settings struct {
	batchInterval time.Duration
}

// Option configures NewWriter.
type Option func(*settings)

// WithBatchInterval overrides DefaultBatchInterval. It must be at least
// cloudwatchwriter2.MinBatchInterval (200ms); NewWriter returns an error
// otherwise.
func WithBatchInterval(d time.Duration) Option {
	return func(s *settings) { s.batchInterval = d }
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
func NewWriter(ctx context.Context, client cloudwatchwriter2.CloudWatchLogsClient, logGroup, streamName string, opts ...Option) (*cloudwatchwriter2.CloudWatchWriter, error) {
	s := settings{batchInterval: DefaultBatchInterval}
	for _, opt := range opts {
		opt(&s)
	}

	// cloudwatchwriter2 uses this ctx for its background flush goroutine and
	// every PutLogEvents call, so a cancelled ctx would silently stop shipping.
	writer, err := cloudwatchwriter2.NewWithClientContext(context.WithoutCancel(ctx), client, s.batchInterval, logGroup, streamName)
	if err != nil {
		return nil, fmt.Errorf("new cloudwatch writer: %w", err)
	}

	return writer, nil
}

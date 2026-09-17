// Package cloudwatchlog builds an io.WriteCloser that ships log lines to
// AWS CloudWatch Logs, suitable for wiring into zerolog (or anything else
// that writes lines to an io.Writer) via zerolog.MultiLevelWriter.
package cloudwatchlog

import (
	"context"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awscloudwatchlogs "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/lzap/cloudwatchwriter2"
)

// DefaultBatchInterval is how often queued log events are flushed to
// CloudWatch when no WithBatchInterval option is given.
const DefaultBatchInterval = 5 * time.Second

type settings struct {
	batchInterval time.Duration
	configOpts    []func(*awsconfig.LoadOptions) error
}

// Option configures NewWriter.
type Option func(*settings)

// WithBatchInterval overrides DefaultBatchInterval. It must be at least
// cloudwatchwriter2.MinBatchInterval (200ms); NewWriter returns an error
// otherwise.
func WithBatchInterval(d time.Duration) Option {
	return func(s *settings) { s.batchInterval = d }
}

// WithConfigOptions passes additional aws-sdk-go-v2 config.LoadOptions to
// the underlying config.LoadDefaultConfig call, e.g. to disable the EC2
// IMDS credential lookup on platforms without an EC2 metadata service.
func WithConfigOptions(opts ...func(*awsconfig.LoadOptions) error) Option {
	return func(s *settings) { s.configOpts = append(s.configOpts, opts...) }
}

// NewWriter builds an io.WriteCloser that ships log lines to logGroup under
// logStreamName, using the ambient AWS config (env vars, shared config/
// credentials files, EC2/ECS/Fly OIDC role, etc). The log group must already
// exist; the log stream is created if it's missing.
//
// Callers on Fly can name the stream with FlyStreamName; other callers
// should pass a name that's unique enough to avoid two processes writing to
// the same stream and stomping on each other's sequence tokens.
func NewWriter(ctx context.Context, logGroup, streamName string, opts ...Option) (*cloudwatchwriter2.CloudWatchWriter, error) {
	s := settings{batchInterval: DefaultBatchInterval}
	for _, opt := range opts {
		opt(&s)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, s.configOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := awscloudwatchlogs.NewFromConfig(awsCfg)

	writer, err := cloudwatchwriter2.NewWithClientContext(ctx, client, s.batchInterval, logGroup, streamName)
	if err != nil {
		return nil, fmt.Errorf("new cloudwatch writer: %w", err)
	}

	return writer, nil
}

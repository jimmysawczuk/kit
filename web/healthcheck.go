package web

import (
	"context"
)

type HealthChecker interface {
	Name() string
	HealthCheck(context.Context) error
}

type HealthCheckerFunc func(context.Context) error

func (h HealthCheckerFunc) HealthCheck(ctx context.Context) error {
	return h(ctx)
}

type namedHealthCheckFunc struct {
	fn   func(context.Context) error
	name string
}

var _ HealthChecker = namedHealthCheckFunc{}

// HealthCheck implements HealthChecker.
func (n namedHealthCheckFunc) HealthCheck(ctx context.Context) error {
	return n.fn(ctx)
}

// Name implements HealthChecker.
func (n namedHealthCheckFunc) Name() string {
	return n.name
}

func NamedHealthCheckFunc(name string, fn HealthCheckerFunc) HealthChecker {
	return namedHealthCheckFunc{
		fn:   fn,
		name: name,
	}
}

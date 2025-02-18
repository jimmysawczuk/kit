package web

type HealthChecker interface {
	Healthy() error
}

type HealthCheck func() error

func (h HealthCheck) Healthy() error {
	return h()
}

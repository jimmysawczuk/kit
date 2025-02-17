package web

// Module is a set of endpoints that can be naturally grouped together. A module can evaluate its own health
// and can attach itself to a Router.
type Module interface {
	Healthy() error
	Route(Router, ...Middleware)
}

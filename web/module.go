package web

import "github.com/jimmysawczuk/kit/web/router"

// Module is a set of endpoints that can be naturally grouped together. A module can evaluate its own health
// and can attach itself to a Router.
type Module interface {
	Route(router.Router)
}

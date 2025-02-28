package web

import (
	"github.com/go-chi/chi"
)

// Module is a set of endpoints that can be naturally grouped together. A module can evaluate its own health
// and can attach itself to a Router.
type Module interface {
	Route(chi.Router, ...Middleware)
}

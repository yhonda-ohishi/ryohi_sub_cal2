package adapters

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
)

// ChiMuxAdapter bridges chi router to gorilla/mux
type ChiMuxAdapter struct {
	muxRouter *mux.Router
	chiRouter chi.Router
}

// NewChiMuxAdapter creates a new adapter to use chi routes in mux
func NewChiMuxAdapter(muxRouter *mux.Router) *ChiMuxAdapter {
	return &ChiMuxAdapter{
		muxRouter: muxRouter,
		chiRouter: chi.NewRouter(),
	}
}

// Mount adds chi routes to mux under the specified prefix
func (a *ChiMuxAdapter) Mount(prefix string, chiSetup func(chi.Router)) {
	// Create a chi router and let the setup function configure it
	chiSetup(a.chiRouter)
	
	// Mount the chi router as a handler under the prefix
	a.muxRouter.PathPrefix(prefix).Handler(http.StripPrefix(prefix, a.chiRouter))
}

// GetChiRouter returns the underlying chi router for direct access
func (a *ChiMuxAdapter) GetChiRouter() chi.Router {
	return a.chiRouter
}

// AdaptChiToMux is a convenience function to add chi routes to a mux router
func AdaptChiToMux(muxRouter *mux.Router, prefix string, chiSetup func(chi.Router)) {
	// Create a new chi router for this mount point
	chiRouter := chi.NewRouter()

	// Let the setup function configure the chi router
	chiSetup(chiRouter)

	// Mount the chi router under the prefix
	muxRouter.PathPrefix(prefix).Handler(http.StripPrefix(prefix, chiRouter))
}
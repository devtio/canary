package routing

import (
	"github.com/gorilla/mux"
)

// NewRouter creates the router with all API routes and the static files handler
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)

	// Build our API server routes and install them.
	routes := NewRoutes()
	for _, route := range routes.Routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return router
}

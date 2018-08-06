package routing

import (
	"net/http"

	"github.com/devtio/canary/handlers"
)

// Route describes a single route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes holds an array of Route
type Routes struct {
	Routes []Route
}

// NewRoutes creates and returns all the API routes
func NewRoutes() (r *Routes) {
	r = new(Routes)

	r.Routes = []Route{
		{
			"ListNamespaces",
			"GET",
			"/api/namespaces",
			handlers.ListNamespaces,
		},
		{
			"ListVirtualServices",
			"GET",
			"/api/virtual-services/{namespace}",
			handlers.ListVirtualServices,
		},
		{
			"CreateVirtualService",
			"POST",
			"/api/virtual-services/{namespace}",
			handlers.CreateVirtualService,
		},
		{
			"ListPods",
			"GET",
			"/api/pods/{namespace}",
			handlers.ListPods,
		},
		{
			"ListPodsByRelease",
			"GET",
			"/api/pods/{namespace}/{release}",
			handlers.ListPods,
		},
		{
			"ListGateways",
			"GET",
			"/api/pods/{namespace}",
			handlers.ListGateways,
		},
		{
			"ListTrafficSegments",
			"GET",
			"/api/traffic-segments/{namespace}",
			handlers.ListTrafficSegments,
		},
		{
			"CreateTrafficSegment",
			"GET",
			"/api/traffic-segments/{namespace}/{releaseId}",
			handlers.CreateTrafficSegment,
		},
	}

	return
}

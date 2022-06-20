package health

import (
	"github.com/gorilla/mux"
	"github.com/resonatecoop/id/util/routes"
)

// RegisterRoutes registers route handlers for the health service
func (s *Service) RegisterRoutes(router *mux.Router, prefix string) {
	subRouter := router.PathPrefix(prefix).Subrouter()
	routes.AddRoutes(s.GetRoutes(), subRouter)
}

// GetRoutes returns []routes.Route slice for the health service
func (s *Service) GetRoutes() []routes.Route {
	return []routes.Route{
		{
			Name:        "health_check",
			Method:      "GET",
			Pattern:     "/health",
			HandlerFunc: s.healthcheck,
		},
	}
}

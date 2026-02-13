package middlewares

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CorsDocs sets CORS headers for the SwaggerDoc UI to be able to access the API.
// TODO: Make this more flexible by allowing the allowed origins to be configured.
func (p *Provider) CorsDocs() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5002"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

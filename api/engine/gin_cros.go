package engine

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowAllOrigins: true,
			//AllowOrigins:    nil,
			AllowMethods: []string{
				"OPTIONS",
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"FETCH",
			},
			AllowHeaders:           []string{"Authorization, Content-Length, X-CSRF-Token, Token,session", "Content-Type", "x-requested-with"},
			AllowCredentials:       true,
			ExposeHeaders:          []string{"Content-Length", "Content-Type"},
			MaxAge:                 86400,
			AllowWildcard:          true,
			AllowBrowserExtensions: true,
			AllowWebSockets:        true,
			AllowFiles:             true,
		},
	)
}

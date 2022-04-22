package web

import (
	"github.com/labstack/echo/v4"
	"github.com/ryboe/q"
)

// AllowOrigin adds some CORS-headers to the response
func AllowOrigin(allowOrigin string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
			q.Q(allowOrigin)
			return next(c)
		}
	}
}

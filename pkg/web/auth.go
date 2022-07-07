package web

import (
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/service"
)

const anyMethod = "*"

// exceptions endpoints
var exceptions = [][2]string{
	{anyMethod, "/login"},
	{anyMethod, "/signup"},
	{anyMethod, "/stop"},
	{anyMethod, "/info"},
	{http.MethodGet, "/jobs"},
	{http.MethodGet, "/jobs/*"},
	{anyMethod, "/notifications"},
}

func authSkip(c echo.Context) bool {
	for _, exc := range exceptions {
		method := c.Request().Method
		if method == http.MethodOptions { // preflight always enabled
			return true
		}

		match, err := path.Match(exc[1], c.Request().RequestURI)
		if err != nil {
			panic(fmt.Errorf("invalid pattern %s: %w", exc, err))
		}

		if match && (exc[0] == anyMethod || exc[0] == method) {
			return true
		}

	}
	return false
}

// Auth checks if there are user authenticated
func Auth(securitySvc service.Security) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !authSkip(c) {
				_, err := securitySvc.FromEchoContext(c)
				if err != nil {
					clog.Ectx(c).Warn().Err(err).Msg("Unable to authenticate user")
					return echo.NewHTTPError(http.StatusUnauthorized, "Authorization required")
				}
			}

			return next(c)
		}
	}
}

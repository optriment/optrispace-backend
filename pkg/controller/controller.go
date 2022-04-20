package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/clog"
)

type (
	// Registerer interface describes registerable controllers
	Registerer interface {
		// Register registers all endpoints in list
		Register(e *echo.Echo)
	}
)

const (
	resourceAuth        = "auth"
	resourceJob         = "jobs"
	resourcePerson      = "persons"
	resourceApplication = "applications"
)

// AddStop adds endpoint to stops application (with cancel function)
func AddStop(e *echo.Echo, cancel context.CancelFunc) {
	e.Any("stop", func(c echo.Context) error {
		cancel()
		clog.Ectx(c).Info().Msg("Exiting...")
		return c.JSON(http.StatusAccepted, echo.HTTPError{Message: "Stop signal accepted"})
	})
}

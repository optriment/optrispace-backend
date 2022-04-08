package controller

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// AddStop adds endpoint to stops application (with cancel function)
func AddStop(e *echo.Echo, cancel context.CancelFunc) {
	e.Any("stop", func(c echo.Context) error {
		cancel()
		log.Info().Msg("Exiting...")
		return c.JSON(http.StatusAccepted, echo.HTTPError{Message: "Stop signal accepted"})
	})
}

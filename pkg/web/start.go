package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
)

// GetErrorHandler creates new error handler. It handles errors in custom way
func GetErrorHandler(oldHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		log.Error().Err(err).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Processing error (see above)")

		switch {
		case errors.Is(err, model.ErrEntityNotFound):
			err = echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
		case errors.Is(err, model.ErrValueIsRequired):
			err = echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Sprintf("Error while processing request: %s", err))
		case errors.Is(err, model.ErrUnauthorized):
			err = echo.NewHTTPError(http.StatusUnauthorized, "Authorization required")
		}

		oldHandler(err, c)
		log.Error().Err(err).Int("status", c.Response().Status).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Failed to process request")
	}
}

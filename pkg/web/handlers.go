package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/model"
)

// GetErrorHandler creates new error handler. It handles errors in custom way
func GetErrorHandler(oldHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		clog.Ectx(c).Error().Err(err).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Processing error (see above)")

		switch {
		case errors.Is(err, model.ErrEntityNotFound):
			err = echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
		case errors.Is(err, model.ErrValueIsRequired):
			err = echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Sprintf("Value is required: %s", err))
		case errors.Is(err, model.ErrUnauthorized):
			err = echo.NewHTTPError(http.StatusUnauthorized, "Authorization required")
		case errors.Is(err, model.ErrDuplication):
			err = echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Duplication: %s", err))
		case errors.Is(err, model.ErrInvalidValue):
			err = echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Sprintf("Invalid value: %s", err))
		}

		oldHandler(err, c)
		clog.Ectx(c).Error().Err(err).Int("status", c.Response().Status).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Failed to process request")
	}
}

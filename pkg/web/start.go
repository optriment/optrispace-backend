package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

// This header will be added to the output log
// Any data connected with request
const HeaderXHint = "X-Hint"

type logger struct{}

var loggerBeacon = logger(struct{}{})

// GetErrorHandler creates new error handler. It handles errors in custom way
func GetErrorHandler(oldHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		EchoLog(c).Error().Err(err).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Processing error (see above)")

		switch {
		case errors.Is(err, model.ErrEntityNotFound):
			err = echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
		case errors.Is(err, model.ErrValueIsRequired):
			err = echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Sprintf("Error while processing request: %s", err))
		case errors.Is(err, model.ErrUnauthorized):
			err = echo.NewHTTPError(http.StatusUnauthorized, "Authorization required")
		case errors.Is(err, model.ErrDuplication):
			err = echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("Duplication: %s", err))
		}

		oldHandler(err, c)
		EchoLog(c).Error().Err(err).Int("status", c.Response().Status).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Failed to process request")
	}
}

// EchoLog returns zerolog.Logger from request the echo.Context
func EchoLog(c echo.Context) *zerolog.Logger {
	return Log(c.Request().Context())
}

// Log returns zerolog.Logger from the execution context
func Log(ctx context.Context) *zerolog.Logger {
	logger, ok := ctx.Value(loggerBeacon).(*zerolog.Logger)
	if !ok {
		panic("inappropriate using logger from context.Context")
	}
	return logger
}

// PrepareContext prepares context.Context for the request
// Adds preconfigured logger with echo.HeaderXRequestID and HeaderXHint (if any) fields
// this middleware should go first in Pre block
func PrepareContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		lc := log.With()
		rid := c.Request().Header.Get(echo.HeaderXRequestID)

		if rid == "" {
			rid = pgdao.NewID()
		}

		lc = lc.Str(echo.HeaderXRequestID, rid)

		if hint := c.Request().Header.Get(HeaderXHint); hint != "" {
			lc = lc.Str(HeaderXHint, hint)
		}

		l := lc.Logger()
		newCtx := context.WithValue(c.Request().Context(), loggerBeacon, &l)
		newReq := c.Request().WithContext(newCtx)
		c.SetRequest(newReq)

		return next(c)
	}
}

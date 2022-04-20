package clog

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HeaderXHint header will be added to the output log
// Any data connected with request
const HeaderXHint = "X-Hint"

type beacon struct{}

var loggerBeacon = beacon(struct{}{})

// NewID generates new unique ID
func NewID() string {
	return shortuuid.New()
}

// Ectx returns zerolog.Logger from request the echo.Context
func Ectx(c echo.Context) *zerolog.Logger {
	return Ctx(c.Request().Context())
}

// Ctx returns zerolog.Logger from the execution context
func Ctx(ctx context.Context) *zerolog.Logger {
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
			rid = NewID()
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

package web

import (
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/pgsvc"
)

// GetErrorHandler creates new error handler. It handles errors in custom way
func GetErrorHandler(oldHandler echo.HTTPErrorHandler) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		clog.Ectx(c).Error().Err(err).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Processing error (see above)")

		var (
			be     *model.BackendError
			status = 0
			cause  error
		)

		switch {
		case errors.Is(err, model.ErrEntityNotFound):
			status = http.StatusNotFound
			cause = model.ErrEntityNotFound

		case errors.Is(err, model.ErrValidationFailed):
			status = http.StatusUnprocessableEntity
			cause = model.ErrValidationFailed

		case errors.Is(err, model.ErrInvalidFormat):
			status = http.StatusBadRequest
			cause = model.ErrInvalidFormat

		case errors.Is(err, model.ErrInsufficientFunds):
			status = http.StatusUnprocessableEntity
			cause = model.ErrInsufficientFunds

		case errors.Is(err, model.ErrUnauthorized):
			status = http.StatusUnauthorized
			cause = model.ErrUnauthorized

		case errors.Is(err, model.ErrDuplication):
			status = http.StatusConflict
			cause = model.ErrDuplication

		case errors.Is(err, model.ErrApplicationAlreadyExists):
			status = http.StatusConflict
			cause = model.ErrApplicationAlreadyExists

		case errors.Is(err, model.ErrUnableToLogin):
			status = http.StatusUnprocessableEntity
			cause = model.ErrUnableToLogin

		case errors.Is(err, model.ErrInappropriateAction):
			status = http.StatusBadRequest
			cause = model.ErrInappropriateAction

		case errors.Is(err, model.ErrInsufficientRights):
			status = http.StatusForbidden
			cause = model.ErrInsufficientRights

		default:
			oldHandler(err, c)
		}

		// sink err to the sentry
		if hub := sentryecho.GetHubFromContext(c); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				if rid := c.Request().Header.Get(echo.HeaderXRequestID); rid != "" {
					scope.SetExtra(echo.HeaderXRequestID, rid)
				}

				if ucv := c.Get(pgsvc.UserContextKey); ucv != nil {
					if uc, ok := ucv.(*model.UserContext); ok {
						scope.SetExtra("user-login", uc.Subject.Login)
						scope.SetExtra("user-id", uc.Subject.ID)
						scope.SetExtra("user-dn", uc.Subject.DisplayName)
					}
				}

				hub.CaptureException(err)
			})
		}

		if status == 0 {
			goto end
		}

		if !errors.As(err, &be) {
			be = &model.BackendError{
				Cause:    cause,
				Message:  cause.Error(),
				TechInfo: err.Error(),
			}
		}

		if e := c.JSON(status, be); e != nil {
			clog.Ectx(c).Warn().Err(e).Msg("Unable to create JSON response for an incoming error")
		}

	end:
		clog.Ectx(c).Error().Err(err).Int("status", c.Response().Status).Stringer("url", c.Request().URL).Str("method", c.Request().Method).
			Msg("Failed to process request")
	}
}

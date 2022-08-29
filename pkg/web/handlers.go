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
// Use this snippet for Swagger specification annotations. You can copy and must fold them with HTTP status.
// I.e. there is only exactly ONE different HTTP status allowed. In Swagger spec will added only last specified HTTP status,
// if this rule not be followed
// @Failure     400  {object} model.BackendError "invalid format"
// @Failure     400  {object} model.BackendError "inappropriate action"
// @Failure     401  {object} model.BackendError "user not authorized"
// @Failure     403  {object} model.BackendError "insufficient rights"
// @Failure     404  {object} model.BackendError "entity not found"
// @Failure     409  {object} model.BackendError "duplication"
// @Failure     409  {object} model.BackendError "application already exists"
// @Failure     422  {object} model.BackendError "validation failed"
// @Failure     422  {object} model.BackendError "insufficient funds"
// @Failure     422  {object} model.BackendError "unable to login"
// @Failure     500  {object} echo.HTTPError{message=string}
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
		case errors.Is(err, model.ErrInvalidFormat):
			status = http.StatusBadRequest // 400
			cause = model.ErrInvalidFormat // "invalid format"

		case errors.Is(err, model.ErrInappropriateAction):
			status = http.StatusBadRequest       // 400
			cause = model.ErrInappropriateAction // "inappropriate action"

		case errors.Is(err, model.ErrUnauthorized):
			status = http.StatusUnauthorized // 401
			cause = model.ErrUnauthorized    // "user not authorized"

		case errors.Is(err, model.ErrInsufficientRights):
			status = http.StatusForbidden       // 403
			cause = model.ErrInsufficientRights // "insufficient rights"

		case errors.Is(err, model.ErrEntityNotFound):
			status = http.StatusNotFound    // 404
			cause = model.ErrEntityNotFound // "entity not found"

		case errors.Is(err, model.ErrDuplication):
			status = http.StatusConflict // 409
			cause = model.ErrDuplication // "duplication"

		case errors.Is(err, model.ErrApplicationAlreadyExists):
			status = http.StatusConflict              // 409
			cause = model.ErrApplicationAlreadyExists // "application already exists"

		case errors.Is(err, model.ErrValidationFailed):
			status = http.StatusUnprocessableEntity // 422
			cause = model.ErrValidationFailed       // "validation failed"

		case errors.Is(err, model.ErrInsufficientFunds):
			status = http.StatusUnprocessableEntity // 422
			cause = model.ErrInsufficientFunds      // "insufficient funds"

		case errors.Is(err, model.ErrUnableToLogin):
			status = http.StatusUnprocessableEntity // 422
			cause = model.ErrUnableToLogin          // "unable to login"

		default:
			oldHandler(err, c) // 500 in most cases
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

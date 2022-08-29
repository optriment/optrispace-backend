package controller

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	// Auth controller
	Auth struct {
		sm     service.Security
		person service.Person
	}
)

// NewAuth create new service
func NewAuth(sm service.Security, person service.Person) Registerer {
	return &Auth{
		sm:     sm,
		person: person,
	}
}

// Register implements Registerer interface
func (cont *Auth) Register(e *echo.Echo) {
	e.POST("/login", cont.login)
	e.POST("/signup", cont.signup)
	e.PUT("/password", cont.newPassword)
	e.GET("/me", cont.me)
	log.Debug().Str("controller", resourceAuth).Msg("Registered")
}

type loginParams struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

// @Summary     Login
// @Description Create user security token for supplied conditionals
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     controller.loginParams true "Login request"
// @Success     200  {object} model.UserContext
// @Failure     422  {object} model.BackendError "unable to login (login or password is not valid)"
// @Failure     500  {object} echo.HTTPError{message=string}
// @Router      /login [post]
func (cont *Auth) login(c echo.Context) error {
	ie := new(loginParams)
	if err := c.Bind(ie); err != nil {
		return err
	}

	ie.Login = strings.ToLower(strings.TrimSpace(ie.Login))

	p, err := cont.sm.FromLoginPassword(c.Request().Context(), ie.Login, ie.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, p)
}

// @Summary     Register a new user
// @Description Register a new user with specified description
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     model.Person true "Register new user"
// @Success     200  {object} model.UserContext
// @Failure     422  {object} echo.HTTPError{message=string} "input object is invalid"
// @Failure     500  {object} echo.HTTPError{message=string}
// @Router      /signup [post]
func (cont *Auth) signup(c echo.Context) error {
	ie := new(model.Person)

	if err := c.Bind(ie); err != nil {
		return err
	}

	ie.Realm = ""

	if ie.Password == "" {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Password required")
	}

	ie.Login = strings.ToLower(strings.TrimSpace(ie.Login))

	o, err := cont.person.Add(c.Request().Context(), ie)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourcePerson, o.ID))

	return c.JSON(http.StatusCreated, model.UserContext{
		Authenticated: true,
		Token:         o.AccessToken,
		Subject:       o,
	})
}

// @Summary     Returns current user information
// @Description Returns information about current authenticated user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Success     200  {object} model.UserContext
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     500  {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /me [get]
func (cont *Auth) me(c echo.Context) error {
	uctx, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return fmt.Errorf("unable to acquire user context: %w", err)
	}

	return c.JSON(http.StatusOK, uctx)
}

type newPasswordParams struct {
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// @Summary     Change password for current authenticated user
// @Description Change password for current authenticated user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body     controller.newPasswordParams true "Change password request"
// @Success     200 {object} model.UserContext
// @Failure     401  {object} model.BackendError "user not authorized or old password is incorrect"
// @Failure     422  {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /password [put]
func (cont *Auth) newPassword(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	ie := new(newPasswordParams)
	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.NewPassword == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("new_password"),
		}
	}

	err = cont.person.UpdatePassword(c.Request().Context(), uc.Subject.ID, ie.OldPassword, ie.NewPassword)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, uc)
}

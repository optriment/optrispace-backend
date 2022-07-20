package controller

import (
	"fmt"
	"net/http"
	"path"

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

func (cont *Auth) login(c echo.Context) error {
	type incoming struct {
		Login    string `json:"login,omitempty"`
		Password string `json:"password,omitempty"`
	}

	ie := new(incoming)
	if err := c.Bind(ie); err != nil {
		return err
	}

	p, err := cont.sm.FromLoginPassword(c.Request().Context(), ie.Login, ie.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, p)
}

func (cont *Auth) signup(c echo.Context) error {
	ie := new(model.Person)

	if err := c.Bind(ie); err != nil {
		return err
	}

	ie.Realm = ""

	if ie.Password == "" {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Password required")
	}

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

func (cont *Auth) me(c echo.Context) error {
	uctx, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return fmt.Errorf("unable to acquire user context: %w", err)
	}

	return c.JSON(http.StatusOK, uctx)
}

func (cont *Auth) newPassword(c echo.Context) error {
	type inputParameters struct {
		OldPassword string `json:"old_password,omitempty"`
		NewPassword string `json:"new_password,omitempty"`
	}

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	ie := new(inputParameters)
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

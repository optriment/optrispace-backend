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
	// Person controller
	Person struct {
		sm  service.Security
		svc service.Person
	}
)

// NewPerson create new service
func NewPerson(sm service.Security, svc service.Person) Registerer {
	return &Person{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Person) Register(e *echo.Echo) {
	e.POST(resourcePerson, cont.add)
	e.GET(resourcePerson, cont.list)
	e.GET(resourcePerson+"/:id", cont.get)
	e.PUT(resourcePerson+"/:id", cont.update)
	log.Debug().Str("controller", resourcePerson).Msg("Registered")
}

// add is NOT signup method. For signup please see /signup endpoint!
func (cont *Person) add(c echo.Context) error {
	input := new(model.Person)

	if err := c.Bind(input); err != nil {
		return err
	}

	input.ID = ""
	o, err := cont.svc.Add(c.Request().Context(), input)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourcePerson, o.ID))
	return c.JSON(http.StatusCreated, o)
}

func (cont *Person) get(c echo.Context) error {
	id := c.Param("id")
	o, err := cont.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

func (cont *Person) list(c echo.Context) error {
	oo, err := cont.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, oo)
}

func (cont *Person) update(c echo.Context) error {
	ie := make(map[string]any)

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	if e := c.Bind(&ie); e != nil {
		return e
	}

	return cont.svc.Patch(c.Request().Context(), id, uc.Subject.ID, ie)
}

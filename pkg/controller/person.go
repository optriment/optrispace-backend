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
		svc service.Person
	}
)

// NewPerson create new service
func NewPerson(svc service.Person) Registerer {
	return &Person{
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Person) Register(e *echo.Echo) {
	e.POST(resourcePerson, cont.add)
	e.GET(resourcePerson, cont.list)
	e.GET(resourcePerson+"/:id", cont.get)
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", resourcePerson).Msg("Registered")
}

func (cont *Person) add(c echo.Context) error {
	person := new(model.Person)

	if err := c.Bind(person); err != nil {
		return err
	}

	person.ID = ""

	person, err := cont.svc.Add(c.Request().Context(), person)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourcePerson, person.ID))
	return c.JSON(http.StatusCreated, person)
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

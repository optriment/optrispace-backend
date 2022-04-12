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
	Person struct {
		name string
		svc  service.Person
	}
)

func NewPerson(svc service.Person) Registerer {
	return &Person{
		name: "persons",
		svc:  svc,
	}
}

func (cont *Person) Register(e *echo.Echo) {
	e.POST(cont.name, cont.add)
	// e.GET(cont.name, cont.list)
	// e.GET(cont.name+"/:id", cont.get)
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", cont.name).Msg("Registered")
}

// nolint: dupl
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

	c.Response().Header().Set("Location", path.Join("/", cont.name, person.ID))
	return c.JSON(http.StatusCreated, person)
}

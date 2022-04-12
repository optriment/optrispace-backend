package controller

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	Job struct {
		name string
		svc  service.Job
	}
)

func NewJob(svc service.Job) Registerer {
	return &Job{
		name: "jobs",
		svc:  svc,
	}
}

func (cont *Job) Register(e *echo.Echo) {
	e.POST(cont.name, cont.add)
	e.GET(cont.name, cont.list)
	e.GET(cont.name+"/:id", cont.get)
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", cont.name).Msg("Registered")
}

// nolint: dupl
func (cont *Job) add(c echo.Context) error {
	job := new(model.Job)

	if err := c.Bind(job); err != nil {
		return err
	}

	job.ID = ""
	job, err := cont.svc.Add(c.Request().Context(), job)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set("Location", path.Join("/", cont.name, job.ID))
	return c.JSON(http.StatusCreated, job)
}

func (cont *Job) list(c echo.Context) error {
	jj, err := cont.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jj)
}

func (cont *Job) get(c echo.Context) error {
	id := c.Param("id")
	job, err := cont.svc.Get(c.Request().Context(), id)
	if errors.Is(model.ErrEntityNotFound, err) {
		return echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
	}

	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, job)
}

package controller

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	// Job controller
	Job struct {
		name string
		svc  service.Job
		sm   service.Security
	}
)

// NewJob create new service
func NewJob(svc service.Job, sm service.Security) Registerer {
	return &Job{
		name: "jobs",
		svc:  svc,
		sm:   sm,
	}
}

// Register implements Registerer interface
func (cont *Job) Register(e *echo.Echo) {
	e.POST(cont.name, cont.add)
	e.GET(cont.name, cont.list)
	e.GET(cont.name+"/:id", cont.get)
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", cont.name).Msg("Registered")
}

func (cont *Job) add(c echo.Context) error {
	type addingJob struct {
		Title       string          `json:"title,omitempty"`
		Description string          `json:"description,omitempty"`
		Budget      decimal.Decimal `json:"budget,omitempty"`
		Duration    int32           `json:"duration,omitempty"`
	}

	uc, err := cont.sm.FromContext(c)
	if err != nil {
		return err
	}

	ae := new(addingJob)

	if e := c.Bind(ae); e != nil {
		return e
	}

	if ae.Title == "" {
		return fmt.Errorf("title required: %w", model.ErrValueIsRequired)
	}

	if ae.Description == "" {
		return fmt.Errorf("title required: %w", model.ErrValueIsRequired)
	}

	o := &model.Job{
		Title:       ae.Title,
		Description: ae.Description,
		Budget:      ae.Budget,
		Duration:    ae.Duration,
		CreatedBy:   uc.Subject,
	}

	o, err = cont.svc.Add(c.Request().Context(), o)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", cont.name, o.ID))
	return c.JSON(http.StatusCreated, o)
}

func (cont *Job) list(c echo.Context) error {
	oo, err := cont.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, oo)
}

func (cont *Job) get(c echo.Context) error {
	id := c.Param("id")
	o, err := cont.svc.Get(c.Request().Context(), id)
	if errors.Is(model.ErrEntityNotFound, err) {
		return echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
	}

	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

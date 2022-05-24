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
		sm  service.Security
		svc service.Job
	}
)

// NewJob create new service
func NewJob(sm service.Security, svc service.Job) Registerer {
	return &Job{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Job) Register(e *echo.Echo) {
	e.POST(resourceJob, cont.add)
	e.GET(resourceJob, cont.list)
	e.GET(resourceJob+"/:id", cont.get)
	e.PUT(resourceJob+"/:id", cont.update)
	log.Debug().Str("controller", resourceJob).Msg("Registered")
}

func (cont *Job) add(c echo.Context) error {
	type addingJob struct {
		Title       string          `json:"title,omitempty"`
		Description string          `json:"description,omitempty"`
		Budget      decimal.Decimal `json:"budget,omitempty"`
		Duration    int32           `json:"duration,omitempty"`
	}

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(addingJob)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.Title == "" {
		return fmt.Errorf("title required: %w", model.ErrValueIsRequired)
	}

	if ie.Description == "" {
		return fmt.Errorf("description required: %w", model.ErrValueIsRequired)
	}

	o := &model.Job{
		Title:       ie.Title,
		Description: ie.Description,
		Budget:      ie.Budget,
		Duration:    ie.Duration,
		CreatedBy:   uc.Subject.ID,
	}

	o, err = cont.svc.Add(c.Request().Context(), o)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceJob, o.ID))
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

func (cont *Job) update(c echo.Context) error {
	ie := make(map[string]any)

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	if e := c.Bind(&ie); e != nil {
		return e
	}

	o, err := cont.svc.Patch(c.Request().Context(), id, uc.Subject.ID, ie)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

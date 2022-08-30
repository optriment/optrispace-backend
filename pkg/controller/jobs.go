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

type jobDescription struct {
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	Budget      decimal.Decimal `json:"budget,omitempty"`
	Duration    int32           `json:"duration,omitempty"`
}

// @Summary     Create a new job
// @Description Creates a new job. Current user will be creator of the job
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       job body     controller.jobDescription true "New job description"
// @Success     200 {object} model.Job
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs [post]
func (cont *Job) add(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(jobDescription)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.Title == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("title"),
		}
	}

	if ie.Description == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("description"),
		}
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

// @Summary     List jobs
// @Description Returns list of jobs
// @Tags        job
// @Accept      json
// @Produce     json
// @Success     200 {array}  model.Job
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs [get]
func (cont *Job) list(c echo.Context) error {
	oo, err := cont.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, oo)
}

// @Summary     Get job description by job_id
// @Description Returns job description by job_id
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       job_id path     string true "Job ID"
// @Success     200    {object} model.Job
// @Failure     404    {object} model.BackendError "job not found"
// @Failure     500    {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{job_id} [get]
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

// @Summary     Update existent job
// @Description Updates existent job by creator only
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       job body     controller.jobDescription true "New job description"
// @Param       id  path     string                    true "Job ID"
// @Success     200 {object} model.Job
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "job not found or current user is not creator"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{id} [put]
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

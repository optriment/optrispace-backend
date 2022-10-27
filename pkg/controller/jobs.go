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
	e.POST(resourceJob+"/:id/block", cont.block)
	e.POST(resourceJob+"/:id/suspend", cont.suspend)
	e.POST(resourceJob+"/:id/resume", cont.resume)
	log.Debug().Str("controller", resourceJob).Msg("Registered")
}

type createJobParams struct {
	Title       string          `json:"title" validate:"required"`
	Description string          `json:"description" validate:"required"`
	Budget      decimal.Decimal `json:"budget"`
	Duration    int32           `json:"duration"`
}

// @Summary     Create a new job
// @Description Creates a new job
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       job body     controller.createJobParams true "Job Params"
// @Success     201 {object} model.JobDTO
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

	ie := new(createJobParams)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if err = validateStruct(ie); err != nil {
		return err
	}

	dto := model.CreateJobDTO{
		Title:       ie.Title,
		Description: ie.Description,
		Budget:      ie.Budget,
		Duration:    ie.Duration,
	}

	newJob, err := cont.svc.Add(c.Request().Context(), uc.Subject.ID, &dto)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceJob, newJob.ID))
	return c.JSON(http.StatusCreated, newJob)
}

// @Summary     List jobs
// @Description Returns list of jobs
// @Tags        job
// @Accept      json
// @Produce     json
// @Success     200 {array}  model.JobDTO
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

// @Summary     Get job by id
// @Description Returns job by id
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       id path string true "Job ID"
// @Success     200 {object} model.JobCardDTO
// @Failure     404 {object} model.BackendError "job not found"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{id} [get]
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

type updateJobParams struct {
	Title       string          `json:"title" validate:"required"`
	Description string          `json:"description" validate:"required"`
	Budget      decimal.Decimal `json:"budget"`
	Duration    int32           `json:"duration"`
}

// @Summary     Update job
// @Description Updates job
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       job body     controller.updateJobParams true "Job params"
// @Param       id  path     string                     true "Job ID"
// @Success     200 {object} model.JobDTO
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "job not found or current user is not creator"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{id} [put]
func (cont *Job) update(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	ie := new(updateJobParams)

	if e := c.Bind(&ie); e != nil {
		return e
	}

	if err = validateStruct(ie); err != nil {
		return err
	}

	dto := model.UpdateJobDTO{
		Title:       ie.Title,
		Description: ie.Description,
		Budget:      ie.Budget,
		Duration:    ie.Duration,
	}

	o, err := cont.svc.Patch(c.Request().Context(), id, uc.Subject.ID, &dto)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

// @Summary     Block a job
// @Description Blocks existent job to hide it from public access. To execute this action, user must have admin privileges.
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       id path string true "Job ID"
// @Success     200
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "user is not admin"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{id}/block [post]
func (cont *Job) block(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	if e := cont.svc.Block(c.Request().Context(), c.Param("id"), uc.Subject.ID); e != nil {
		return e
	}
	return c.JSON(http.StatusOK, "{}")
}

// @Summary     Suspend a job
// @Description Suspends existent job to stop receiving applications
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Job ID"
// @Success     200
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "user is not an owner"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{id}/suspend [post]
func (cont *Job) suspend(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	if e := cont.svc.Suspend(c.Request().Context(), c.Param("id"), uc.Subject.ID); e != nil {
		return e
	}
	return c.JSON(http.StatusOK, "{}")
}

// @Summary     Resume a job
// @Description Resumes existent job to continue receiving applications
// @Tags        job
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Job ID"
// @Success     200
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "user is not an owner"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{id}/resume [post]
func (cont *Job) resume(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	if e := cont.svc.Resume(c.Request().Context(), c.Param("id"), uc.Subject.ID); e != nil {
		return e
	}
	return c.JSON(http.StatusOK, "{}")
}

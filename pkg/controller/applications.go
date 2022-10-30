package controller

import (
	"encoding/json"
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
	// Application controller
	Application struct {
		sm  service.Security
		svc service.Application
	}
)

// NewApplication create new service
func NewApplication(sm service.Security, svc service.Application) Registerer {
	return &Application{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Application) Register(e *echo.Echo) {
	e.POST(resourceJob+"/:job_id/"+resourceApplication, cont.add)
	e.GET(resourceApplication, cont.listByApplicant)
	e.GET(resourceApplication+"/:id", cont.get)
	e.GET(resourceApplication+"/:id/chat", cont.getChat)
	e.GET(resourceJob+"/:job_id/application", cont.getForJob)
	e.GET(resourceJob+"/:job_id/"+resourceApplication, cont.listByJob)
	log.Debug().Str("controller", resourceApplication).Msg("Registered")
}

type createApplicationParams struct {
	Comment string          `json:"comment" validate:"required"`
	Price   decimal.Decimal `json:"price" validate:"required"`
}

// @Summary     Creates a new application for a job
// @Description Applicant creates a new application for a job
// @Tags        application, job
// @Accept      json
// @Produce     json
// @Param       application body     controller.createApplicationParams true "New application request"
// @Param       job_id      path     string                             true "Job ID to apply"
// @Success     201         {object} model.ApplicationDTO
// @Failure     401         {object} model.BackendError "user not authorized"
// @Failure     404         {object} model.BackendError "job with specified ID is not found"
// @Failure     422         {object} model.BackendError "validation failed"
// @Failure     500         {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{job_id}/applications [post]
func (cont *Application) add(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(createApplicationParams)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if err = validateStruct(ie); err != nil {
		return err
	}

	dto := model.CreateApplicationDTO{
		JobID:   c.Param("job_id"),
		Comment: ie.Comment,
		Price:   ie.Price,
	}

	newApplication, err := cont.svc.Add(c.Request().Context(), uc.Subject.ID, &dto)
	if err != nil {
		return fmt.Errorf("unable to add application: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceApplication, newApplication.ID))
	return c.JSON(http.StatusCreated, newApplication)
}

// @Summary     Get an application
// @Description Returns an application by ID
// @Tags        application
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Application ID"
// @Success     200    {object} model.ApplicationDTO
// @Failure     401    {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "application not found"
// @Failure     500    {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /applications/{id} [get]
func (cont *Application) get(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Get(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

// @Summary     Get an application for specific job and applicant
// @Description Returns an application
// @Tags        application
// @Accept      json
// @Produce     json
// @Param       job_id path     string true "Job ID"
// @Success     200 {object} model.ApplicationDTO
// @Failure     401    {object} model.BackendError "user not authorized"
// @Failure     404    {object} model.BackendError "job not found"
// @Failure     500    {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{job_id}/application [get]
func (cont *Application) getForJob(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.GetForJob(c.Request().Context(), c.Param("job_id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	// In case of applicant has not applied yet
	if o == nil {
		return c.JSON(http.StatusOK, json.RawMessage("{}"))
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Get chat for an application
// @Description Performer or customer is getting chat for this application
// @Tags        application,chat
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Application ID"
// @Success     200 {object} model.Chat
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "application not found or user has no permissions for view chat"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /applications/{id}/chat [get]
func (cont *Application) getChat(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	id := c.Param("id")
	o, err := cont.svc.GetChat(ctx, id, uc.Subject.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

// @Summary     List applications for the job
// @Description Returns applications list for the job by job_id
// @Tags        application, job
// @Accept      json
// @Produce     json
// @Param       job_id path     string true "Job ID"
// @Success     200    {array}  model.ApplicationDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404    {object} model.BackendError "job not found"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{job_id}/applications [get]
func (cont *Application) listByJob(c echo.Context) error {
	ctx := c.Request().Context()
	jobID := c.Param("job_id")

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	oo, err := cont.svc.ListByJob(ctx, jobID, uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, oo)
}

// @Summary     List of applications
// @Description Returns applications
// @Tags        application
// @Accept      json
// @Produce     json
// @Success     200 {array}  model.ApplicationDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /applications [get]
func (cont *Application) listByApplicant(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	oo, err := cont.svc.ListByApplicant(c.Request().Context(), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, oo)
}

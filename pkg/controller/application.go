package controller

import (
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
	e.GET(resourceApplication+"/:id", cont.get)
	e.GET(resourceApplication+"/my", cont.listMy)
	e.GET(resourceApplication, cont.list)
	e.GET(resourceJob+"/:job_id/"+resourceApplication, cont.list)
	log.Debug().Str("controller", resourceApplication).Msg("Registered")
}

type newApplication struct {
	Comment string          `json:"comment,omitempty"`
	Price   decimal.Decimal `json:"price,omitempty"`
}

// @Summary     Make new application for a job
// @Description Applicant create new application for a job
// @Tags        application, job
// @Accept      json
// @Produce     json
// @Param       application body     controller.newApplication true "New application request"
// @Param       job_id      path     string                    true "Job ID to apply"
// @Success     200         {object} model.Application
// @Failure     401         {object} model.BackendError "user not authorized"
// @Failure     404         {object} model.BackendError "job with specified ID is not found"
// @Failure     422         {object} model.BackendError "validation failed (details in response)"
// @Failure     500         {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{job_id}/applications [post]
func (cont *Application) add(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(newApplication)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.Comment == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("comment"),
		}
	}

	if decimal.Zero.Equal(ie.Price) {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("price"),
		}
	}

	createdAppl, err := cont.svc.Add(c.Request().Context(), &model.Application{
		Applicant: uc.Subject,
		Comment:   ie.Comment,
		Price:     ie.Price,
		Job:       &model.Job{ID: c.Param("job_id")},
	})
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceApplication, createdAppl.ID))
	return c.JSON(http.StatusCreated, createdAppl)
}

// @Summary     Get an application
// @Description Applicant create new application for a job
// @Tags        application
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Application ID"
// @Success     200 {object} model.Application
// @Failure     401    {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "application not found"
// @Failure     500    {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /applications/{id} [get]
func (cont *Application) get(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	o, err := cont.svc.Get(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

// @Summary     List applications by current authenticated user
// @Description Returns applications list for the current user — as an applicant or as a job creator
// @Tags        application
// @Accept      json
// @Produce     json
// @Success     200 {array}  model.Application
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /applications [get]
func _() {}

// @Summary     List applications for the job
// @Description Returns applications list for the job by job_id. The current user MUST be an applicant or a job creator.
// @Tags        application, job
// @Accept      json
// @Produce     json
// @Param       job_id path     string true "Job ID"
// @Success     200    {array}  model.Application
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404    {object} model.BackendError "job not found"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /jobs/{job_id}/applications [get]
func (cont *Application) list(c echo.Context) error {
	ctx := c.Request().Context()
	jobID := c.Param("job_id")

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	if jobID == "" { // full list for the current user
		oo, e := cont.svc.ListBy(ctx, "", uc.Subject.ID)
		if e != nil {
			return e
		}
		return c.JSON(http.StatusOK, oo)
	}

	oo, err := cont.svc.ListBy(ctx, jobID, uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, oo)
}

// @Summary     List applications were applied by the current authenticated user
// @Description Returns applications list made by the current user as an applicant
// @Tags        application
// @Accept      json
// @Produce     json
// @Success     200 {array}  model.Application
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /applications/my [get]
func (cont *Application) listMy(c echo.Context) error {
	ctx := c.Request().Context()

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	oo, err := cont.svc.ListByApplicant(ctx, uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, oo)
}

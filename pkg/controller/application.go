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
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", resourceApplication).Msg("Registered")
}

func (cont *Application) add(c echo.Context) error {
	type incomingApplication struct {
		Comment string          `json:"comment,omitempty"`
		Price   decimal.Decimal `json:"price,omitempty"`
	}

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(incomingApplication)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.Comment == "" {
		return fmt.Errorf("comment required: %w", model.ErrValueIsRequired)
	}

	if decimal.Zero.Equal(ie.Price) {
		return fmt.Errorf("price required: %w", model.ErrValueIsRequired)
	}

	createdAppl, err := cont.svc.Add(c.Request().Context(), &model.Application{
		Applicant: uc.Subject,
		Comment:   ie.Comment,
		Price:     ie.Price,
		Job:       &model.Job{ID: c.Param("job_id")},
	})
	if err != nil {
		return fmt.Errorf("unable to create application %+v: %w", ie, err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceApplication, createdAppl.ID))
	return c.JSON(http.StatusCreated, createdAppl)
}

func (cont *Application) get(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	o, err := cont.svc.Get(ctx, id)
	if errors.Is(model.ErrEntityNotFound, err) {
		return echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
	}

	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

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

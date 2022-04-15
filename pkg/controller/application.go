package controller

import (
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/service"
)

type (
	// Application controller
	Application struct {
		name string
		svc  service.Application
	}
)

// NewApplication create new service
func NewApplication(svc service.Application) Registerer {
	return &Application{
		name: "applications",
		svc:  svc,
	}
}

// Register implements Registerer interface
func (cont *Application) Register(e *echo.Echo) {
	e.POST(cont.name, cont.add)
	// e.GET(name, cont.list)
	// e.GET(name+"/:id", cont.get)
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", cont.name).Msg("Registered")
}

func (cont *Application) add(c echo.Context) error {
	type appl struct {
		ApplicantID string `json:"applicant_id,omitempty"`
		JobID       string `json:"job_id,omitempty"`
	}

	application := new(appl)

	if err := c.Bind(application); err != nil {
		return err
	}

	createdAppl, err := cont.svc.Add(c.Request().Context(), application.ApplicantID, application.JobID)
	if err != nil {
		return fmt.Errorf("unable to create application %+v: %w", application, err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", cont.name, createdAppl.ID))
	return c.JSON(http.StatusCreated, createdAppl)
}

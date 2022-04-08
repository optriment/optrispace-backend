package controller

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	JobController struct {
		jsvc service.Job
	}
)

func NewJobController(jsvc service.Job) Registerer {
	return &JobController{jsvc: jsvc}
}

func (cont *JobController) Register(e *echo.Echo) {
	e.POST("jobs", cont.create)
	e.GET("jobs", cont.list)
	e.GET("jobs/:id", cont.view)
	// e.PUT("jobs/:id", cont.update)
	log.Debug().Str("controller", "job").Msg("Registered")
}

func (cont *JobController) create(c echo.Context) error {
	job := new(model.Job)

	if err := c.Bind(job); err != nil {
		return err
	}

	job.ID = ""
	cont.jsvc.Save(c.Request().Context(), job)

	c.Response().Header().Set("Location", "/jobs/"+job.ID)
	return c.JSON(http.StatusCreated, job)
}

func (cont *JobController) list(c echo.Context) error {
	jj, err := cont.jsvc.ReadAll(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jj)
}

func (cont *JobController) view(c echo.Context) error {
	id := c.Param("id")
	job, err := cont.jsvc.Read(c.Request().Context(), id)
	if errors.Is(model.ErrEntityNotFound, err) {
		return echo.NewHTTPError(http.StatusNotFound, "Entity with specified id not found")
	}

	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, job)
}

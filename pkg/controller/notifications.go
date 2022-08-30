package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	// Notification controller
	Notification struct {
		svc service.Notification
	}
)

// NewNotification create new service
func NewNotification(svc service.Notification) Registerer {
	return &Notification{
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Notification) Register(e *echo.Echo) {
	e.POST(resourceNotification, cont.add)
	// e.PUT(name+"/:id", cont.update)
	log.Debug().Str("controller", resourceNotification).Msg("Registered")
}

// Warning! We should NOT add this method to the Swagger specification
func (cont *Notification) add(c echo.Context) error {
	bb, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %w", model.ErrInvalidFormat)
	}

	err = cont.svc.Push(c.Request().Context(), string(bb))
	if err != nil {
		return err
	}
	return c.JSONBlob(http.StatusOK, []byte("{}"))
}

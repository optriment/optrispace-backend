package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/service"
)

type (
	// Stats controller
	Stats struct {
		sm  service.Security
		svc service.Stats
	}
)

// NewStats create new service
func NewStats(sm service.Security, svc service.Stats) Registerer {
	return &Stats{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Stats) Register(e *echo.Echo) {
	e.GET(resourceStats, cont.stats)
	log.Debug().Str("controller", resourceStats).Msg("Registered")
}

func (cont *Stats) stats(c echo.Context) error {
	o, err := cont.svc.Stats(c.Request().Context())
	if err != nil {
		return fmt.Errorf("unable to collect stats: %w", err)
	}

	return c.JSONPretty(http.StatusOK, o, "  ")
}

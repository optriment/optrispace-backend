package controller

import "github.com/labstack/echo/v4"

type (
	// Registerer interface describes registerable controllers
	Registerer interface {
		// Register registers all endpoints in list
		Register(e *echo.Echo)
	}
)

const (
	resourceJob         = "jobs"
	resourcePerson      = "persons"
	resourceApplication = "applications"
)
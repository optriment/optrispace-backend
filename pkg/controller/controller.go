package controller

import "github.com/labstack/echo/v4"

type (
	Registerer interface {
		// Register registers all endpoints in list
		Register(e *echo.Echo)
	}
)

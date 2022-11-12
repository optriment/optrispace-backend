package controller

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// In this file we should specify common API infos

// @title       OptriSpace API
// @version     1.0
// @description OptriSpace Server API

// @license.name MIT
// @license.url  https://github.com/optriment/optrispace-backend/blob/develop/LICENSE

// @host     localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerToken
// @in                         header
// @name                       Authorization
// @description                Bearer token in Authorization header

// SwaggerRegister publishes swagger specification with /swagger/index.html
func SwaggerRegister(e *echo.Echo) {
	e.GET("/swagger/*", echoSwagger.WrapHandler)
}

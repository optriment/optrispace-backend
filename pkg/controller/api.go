package controller

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// In this file we should specify common API infos

// @title       Optrispace API
// @version     1.0
// @description Optrispace Server API

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

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

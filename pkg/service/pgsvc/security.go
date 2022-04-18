package pgsvc

import (
	"database/sql"
	"strings"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/web"
)

type (
	// SecuritySvc service
	SecuritySvc struct {
		db *sql.DB
	}
)

// NewSecurity creates new security service
func NewSecurity(db *sql.DB) *SecuritySvc {
	return &SecuritySvc{
		db: db,
	}
}

// BearerPrefix represents suffix for bearer authentication
const BearerPrefix = "Bearer "

// FromContext implements interface SecurityManager
func (s *SecuritySvc) FromContext(c echo.Context) (*model.UserContext, error) {
	prefixLen := len(BearerPrefix)
	auth := c.Request().Header.Get(echo.HeaderAuthorization)

	if len(auth) > prefixLen && strings.EqualFold(BearerPrefix, auth[0:prefixLen]) {
		auth = auth[prefixLen:]
	}

	auth = strings.TrimSpace(auth)

	p, err := NewPerson(s.db).Get(c.Request().Context(), auth)
	if err != nil {
		web.EchoLog(c).Warn().Err(err).Msg("Unable to authorize")
		err = model.ErrUnauthorized
	}

	return &model.UserContext{
		Authorized: err == nil,
		Subject:    p,
	}, err
}

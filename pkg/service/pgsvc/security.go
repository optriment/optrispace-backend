package pgsvc

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// SecuritySvc service
	SecuritySvc struct {
		db *sql.DB
	}
)

const (
	// BearerPrefix represents suffix for bearer authentication
	BearerPrefix = "Bearer "
)

// UserContextKey means user information in the context
const UserContextKey = "user-context"

// NewSecurity creates new security service
func NewSecurity(db *sql.DB) *SecuritySvc {
	return &SecuritySvc{
		db: db,
	}
}

// FromEchoContext implements interface SecurityManager
// It modifies echo.Context!!!
func (s *SecuritySvc) FromEchoContext(c echo.Context) (*model.UserContext, error) {
	exist := c.Get(UserContextKey)
	if exist != nil {
		if uctx, ok := exist.(*model.UserContext); ok {
			return uctx, nil // is already acquired
		}
	}

	prefixLen := len(BearerPrefix)
	auth := c.Request().Header.Get(echo.HeaderAuthorization)

	if len(auth) > prefixLen && strings.EqualFold(BearerPrefix, auth[:prefixLen]) {
		auth = auth[prefixLen:]
	}

	token := strings.TrimSpace(auth)

	// here is token used as is
	personSvc := NewPerson(s.db)
	p, err := personSvc.GetByAccessToken(c.Request().Context(), token)
	if err != nil {
		clog.Ectx(c).Warn().Err(err).Msg("Unable to authorize")
		err = model.ErrUnauthorized
	}

	newUctx := &model.UserContext{
		Authenticated: err == nil,
		Token:         token,
		Subject:       p,
	}

	c.Set(UserContextKey, newUctx)
	return newUctx, err
}

// FromLoginPassword implements service.Security
func (s *SecuritySvc) FromLoginPassword(ctx context.Context, login, password string) (*model.UserContext, error) {
	newUctx := new(model.UserContext)
	return newUctx, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		u, err := queries.PersonGetByLogin(ctx, pgdao.PersonGetByLoginParams{
			Login: login,
			Realm: model.InhouseRealm,
		})

		if errors.Is(err, sql.ErrNoRows) {
			clog.Ctx(ctx).Warn().Str("login", login).Msg("No such login")
			return model.ErrUnableToLogin
		}

		if err != nil {
			return err
		}

		if CompareHashAndPassword(u.PasswordHash, password) != nil {
			clog.Ctx(ctx).Warn().Str("login", login).Msg("Invalid password")
			return model.ErrUnableToLogin
		}

		token := pgdao.NewID()
		if e := queries.PersonSetAccessToken(ctx, pgdao.PersonSetAccessTokenParams{
			AccessToken: token,
			ID:          u.ID,
		}); e != nil {
			clog.Ctx(ctx).Warn().
				Str("login", login).
				Str("id", u.ID).
				Str("token", token).
				Msg("Unable to set access token for user")
			return e
		}

		newUctx.Authenticated = true
		newUctx.Token = token
		newUctx.Subject = &model.Person{
			ID:              u.ID,
			Login:           u.Login,
			DisplayName:     u.DisplayName,
			CreatedAt:       u.CreatedAt,
			Email:           u.Email,
			IsAdmin:         u.IsAdmin,
			EthereumAddress: u.EthereumAddress,
		}
		return nil
	})
}

// CreateHashFromPassword evaluates password hash and returns it as Base64 string
func CreateHashFromPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		panic(fmt.Errorf("unable to evaluate password hash: %w", err))
	}
	return base64.StdEncoding.EncodeToString(hash)
}

// CompareHashAndPassword compares password (UTF-8 encoded string) against hash (base64 encoded string)
// returns nil if password conforms to hash
func CompareHashAndPassword(base64hash, password string) error {
	hash, err := base64.StdEncoding.DecodeString(base64hash)
	if err != nil {
		return fmt.Errorf("unable to decode string %s into hash: %w", base64hash, err)
	}
	return bcrypt.CompareHashAndPassword(hash, []byte(password))
}

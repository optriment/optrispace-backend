package intest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/pgsvc"
)

func TestAuth(t *testing.T) {
	signupURL := appURL + "/signup"
	loginURL := appURL + "/login"
	meURL := appURL + "/me"

	require.NoError(t, pgdao.PurgeDB(bgctx, db))

	var me *model.UserContext

	t.Run("signup•full", func(t *testing.T) {
		body := `{
			"login":"mylogin",
			"password":"12345678",
			"display_name": "John Smith"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.UserContext)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			me = e

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/persons/"+e.Subject.ID))

			assert.True(t, e.Authenticated)
			assert.Equal(t, e.Token, e.Subject.ID)
			if assert.NotNil(t, e.Subject) {
				assert.NotEmpty(t, e.Subject.ID)
				assert.Equal(t, "mylogin", e.Subject.Login)
				assert.Equal(t, "inhouse", e.Subject.Realm)
				assert.Equal(t, "John Smith", e.Subject.DisplayName)
				assert.NotEmpty(t, e.Subject.CreatedAt)
			}

			d, err := pgdao.New(db).PersonGet(bgctx, e.Subject.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.Subject.ID, d.ID)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, "mylogin", d.Login)
				assert.NoError(t, pgsvc.CompareHashAndPassword(d.PasswordHash, "12345678"))
				assert.Equal(t, "John Smith", d.DisplayName)
				assert.Equal(t, e.Subject.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, "", d.Email)
			}
		}
	})

	t.Run("signup•only-password", func(t *testing.T) {
		body := `{
			"password":"12345678"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.UserContext)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/persons/"+e.Subject.ID))

			assert.True(t, e.Authenticated)
			assert.Equal(t, e.Token, e.Subject.ID)
			if assert.NotNil(t, e.Subject) {
				assert.NotEmpty(t, e.Subject.ID)
				assert.Equal(t, e.Subject.ID, e.Subject.Login)
				assert.Equal(t, "inhouse", e.Subject.Realm)
				assert.NotEmpty(t, e.Subject.DisplayName)
				assert.NotEmpty(t, e.Subject.CreatedAt)
			}

			d, err := pgdao.New(db).PersonGet(bgctx, e.Subject.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.Subject.ID, d.ID)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, e.Subject.ID, d.Login)
				assert.NoError(t, pgsvc.CompareHashAndPassword(d.PasswordHash, "12345678"))
				assert.Equal(t, e.Subject.DisplayName, d.DisplayName)
				assert.Equal(t, e.Subject.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, "", d.Email)
			}
		}
	})
	t.Run("signup•wo-password", func(t *testing.T) {
		body := `{
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Password required", e["message"])
		}
	})

	t.Run("me•401", func(t *testing.T) {
		body := `{
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, meURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("me•ok", func(t *testing.T) {
		body := `{
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, meURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+me.Subject.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.UserContext)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, e.Authenticated)
			assert.Equal(t, e.Token, e.Subject.ID)
			if assert.NotNil(t, e.Subject) {
				assert.NotEmpty(t, e.Subject.ID)
				assert.Equal(t, "mylogin", e.Subject.Login)
				assert.Equal(t, "inhouse", e.Subject.Realm)
				assert.Equal(t, "John Smith", e.Subject.DisplayName)
				assert.NotEmpty(t, e.Subject.CreatedAt)
			}

			d, err := pgdao.New(db).PersonGet(bgctx, e.Subject.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.Subject.ID, d.ID)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, "mylogin", d.Login)
				assert.NoError(t, pgsvc.CompareHashAndPassword(d.PasswordHash, "12345678"))
				assert.Equal(t, "John Smith", d.DisplayName)
				assert.Equal(t, e.Subject.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, "", d.Email)
			}
		}
	})

	t.Run("login•ok", func(t *testing.T) {
		body := `{
				"login":"mylogin",
				"password":"12345678"
			}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, loginURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.UserContext)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, e.Authenticated)
			assert.Equal(t, e.Token, e.Subject.ID)
			if assert.NotNil(t, e.Subject) {
				assert.NotEmpty(t, e.Subject.ID)
				assert.Equal(t, "mylogin", e.Subject.Login)
				assert.Equal(t, "inhouse", e.Subject.Realm)
				assert.Equal(t, "John Smith", e.Subject.DisplayName)
				assert.NotEmpty(t, e.Subject.CreatedAt)
			}

			d, err := pgdao.New(db).PersonGet(bgctx, e.Subject.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.Subject.ID, d.ID)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, "mylogin", d.Login)
				assert.NoError(t, pgsvc.CompareHashAndPassword(d.PasswordHash, "12345678"))
				assert.Equal(t, "John Smith", d.DisplayName)
				assert.Equal(t, e.Subject.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, "", d.Email)
			}
		}
	})

	t.Run("login•invalid-password", func(t *testing.T) {
		body := `{
				"login":"mylogin",
				"password":"------"
			}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, loginURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.UserContext)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.False(t, e.Authenticated)
			assert.Empty(t, e.Token)
			assert.Empty(t, e.Subject)
		}
	})

	t.Run("login•invalid-login", func(t *testing.T) {
		body := `{
				"login": "far-long-invalid-login",
				"password": ""
			}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, loginURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.UserContext)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.False(t, e.Authenticated)
			assert.Empty(t, e.Token)
			assert.Empty(t, e.Subject)
		}
	})
}

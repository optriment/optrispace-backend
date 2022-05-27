package intest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jaswdr/faker"
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

	require.NoError(t, pgdao.PurgeDB(ctx, db))

	var me *model.UserContext

	t.Run("signup•full", func(t *testing.T) {
		body := `{
			"login":"mylogin",
			"password":"12345678",
			"display_name": "John Smith"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
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

			d, err := pgdao.New(db).PersonGet(ctx, e.Subject.ID)
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

	t.Run("signup•duplication login", func(t *testing.T) {
		body := `{
			"login":"mylogin",
			"password":"abcde",
			"display_name": "` + faker.New().Person().Name() + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusConflict, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "duplication", e.Message)
		}
	})

	t.Run("signup•only-password", func(t *testing.T) {
		body := `{
			"password":"12345678"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
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

			d, err := pgdao.New(db).PersonGet(ctx, e.Subject.ID)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, signupURL, bytes.NewReader([]byte(body)))
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, meURL, bytes.NewReader([]byte(body)))
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, meURL, bytes.NewReader([]byte(body)))
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

			d, err := pgdao.New(db).PersonGet(ctx, e.Subject.ID)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader([]byte(body)))
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

			d, err := pgdao.New(db).PersonGet(ctx, e.Subject.ID)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader([]byte(body)))
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader([]byte(body)))
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

func TestChangePassword(t *testing.T) {
	ctx := context.Background()

	passwordURL := appURL + "/password"

	require.NoError(t, pgdao.PurgeDB(ctx, db))
	queries := pgdao.New(db)

	t.Run("ok", func(t *testing.T) {
		smith, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			DisplayName:  "Smith 1234",
			Email:        "smith@sample.com",
		})
		require.NoError(t, err)

		body := `{
				"old_password": "1234",
				"new_password": "abcd"
			}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, passwordURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+smith.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Empty(t, bb)
			p, err := queries.PersonGet(ctx, smith.ID)
			if assert.NoError(t, err) {
				assert.NoError(t, pgsvc.CompareHashAndPassword(p.PasswordHash, "abcd"), "Password has to be changed")
			}
		}
	})

	t.Run("invalid old password", func(t *testing.T) {
		smith, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			DisplayName:  "Smith 1234",
			Email:        "smith@sample.com",
		})
		require.NoError(t, err)

		body := `{
				"old_password": "0987",
				"new_password": "abcd"
			}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, passwordURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+smith.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "user not authorized", e.Message)
		}
	})

	t.Run("invalid authorization token", func(t *testing.T) {
		smith, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			DisplayName:  "Smith 1234",
			Email:        "smith@sample.com",
		})
		require.NoError(t, err)

		body := `{
				"old_password": "1234",
				"new_password": "abcd"
			}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, passwordURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+smith.ID+"bad tail")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("empty new password", func(t *testing.T) {
		smith, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			DisplayName:  "Smith 1234",
			Email:        "smith@sample.com",
		})
		require.NoError(t, err)

		body := `{
				"old_password": "1234"
			}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, passwordURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+smith.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "new_password: is required", e.Message)
		}
	})
}

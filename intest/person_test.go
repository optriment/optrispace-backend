package intest

import (
	"bytes"
	"encoding/json"
	"io"
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

func TestPerson(t *testing.T) {
	resourceName := "persons"
	startURL := appURL + "/" + resourceName

	var createdID string

	require.NoError(t, pgdao.PurgeDB(ctx, db))

	creator, err := pgsvc.NewPerson(db).Add(ctx, &model.Person{
		Password: "12345678",
	})
	require.NoError(t, err)

	t.Run("post full", func(t *testing.T) {
		body := `{
			"realm": "my-realm",
			"login": "user1",
			"password": "12345678",
			"display_name": "Breanne McGlynn",
			"email":"predovic.macy@hotmail.com"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Person)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			createdID = e.ID

			assert.True(t, strings.HasPrefix(res.Header.Get("location"), "/"+resourceName+"/"+e.ID))
			assert.NotEmpty(t, e.ID)
			assert.Equal(t, "user1", e.Login)
			assert.Equal(t, "my-realm", e.Realm)
			assert.Equal(t, "Breanne McGlynn", e.DisplayName)
			assert.NotEmpty(t, e.CreatedAt)
			assert.Equal(t, "predovic.macy@hotmail.com", e.Email)

			d, err := pgdao.New(db).PersonGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, "my-realm", d.Realm)
				assert.Equal(t, "user1", d.Login)
				assert.NoError(t, pgsvc.CompareHashAndPassword(d.PasswordHash, "12345678"))
				assert.Equal(t, "Breanne McGlynn", d.DisplayName)
				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, "predovic.macy@hotmail.com", d.Email)
			}
		}
	})

	t.Run("get list", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Person, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			if assert.NotEmpty(t, ee) {
				for _, e := range ee {
					assert.NotEmpty(t, e.ID)
					assert.NotEmpty(t, e.CreatedAt)
				}
			}
		}
	})

	t.Run("get by id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL+"/"+createdID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Person)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, createdID, e.ID)

			assert.NotEmpty(t, e.ID)
			assert.Equal(t, "user1", e.Login)
			assert.Equal(t, "my-realm", e.Realm)
			assert.Equal(t, "Breanne McGlynn", e.DisplayName)
			assert.NotEmpty(t, e.CreatedAt)
			assert.Equal(t, "predovic.macy@hotmail.com", e.Email)
			assert.EqualValues(t, "{}", e.Resources)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL+"/not-existent-entity", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("get by id not authorized", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL+"/not-existent-entity", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})
}

func TestPersonPatch(t *testing.T) {
	personURL := appURL + "/persons"

	require.NoError(t, pgdao.PurgeDB(ctx, db))
	queries := pgdao.New(db)

	t.Run("patch ethereum_address ok", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Empty(t, bb)
			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "0x1234567890abcd", d.EthereumAddress)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, thePerson.Login, d.Login)
				assert.NotEmpty(t, d.PasswordHash)
				assert.Equal(t, "John Smith", d.DisplayName)
				assert.Equal(t, "js@sample.com", d.Email)
			}
		}
	})

	t.Run("patch display_name ok", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"display_name":"Jude Law"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Empty(t, bb)
			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "", d.EthereumAddress)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, thePerson.Login, d.Login)
				assert.NotEmpty(t, d.PasswordHash)
				assert.Equal(t, "Jude Law", d.DisplayName)
				assert.Equal(t, "js@sample.com", d.Email)
			}
		}
	})

	t.Run("patch invalid person", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		theStranger, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			DisplayName:  "Stranger",
			Email:        "s@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+theStranger.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "insufficient rights", e.Message)
		}
	})

	t.Run("patch not authorized", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
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

	t.Run("patch not found", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+pgdao.NewID(), bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "insufficient rights", e.Message)
		}
	})
}

func TestPersonResources(t *testing.T) {
	personURL := appURL + "/persons"

	require.NoError(t, pgdao.PurgeDB(ctx, db))
	queries := pgdao.New(db)

	t.Run("put resources to new person", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"GitHub": "https://github.com/almaz-uno",
			 "Telegram messenger": "https://t.me/almaz_develop_bot"
		}`

		refinedBody := strings.NewReplacer("\t", "", "\n", "").Replace(body)

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.EqualValues(t, "{}\n", string(bb))
			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "", d.EthereumAddress)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, thePerson.Login, d.Login)
				assert.NotEmpty(t, d.PasswordHash)
				assert.Equal(t, "John Smith", d.DisplayName)
				assert.Equal(t, "js@sample.com", d.Email)
				assert.EqualValues(t, refinedBody, string(d.Resources))
			}
		}

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, personURL+"/"+thePerson.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			p := new(model.Person)
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			err = json.Unmarshal(bb, p)
			require.NoError(t, err)

			assert.EqualValues(t, refinedBody, p.Resources)
		}
	})

	t.Run("put resources to none empty", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		err = queries.PersonSetResources(ctx, pgdao.PersonSetResourcesParams{
			Resources: []byte(`{"field":"some value"}`),
			ID:        thePerson.ID,
		})
		require.NoError(t, err)

		body := `{
			"GitHub": "https://github.com/almaz-uno",
			 "Telegram messenger": "https://t.me/almaz_develop_bot"
		}`

		refinedBody := strings.NewReplacer("\t", "", "\n", "").Replace(body)

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.EqualValues(t, "{}\n", string(bb))
			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "", d.EthereumAddress)
				assert.Equal(t, "inhouse", d.Realm)
				assert.Equal(t, thePerson.Login, d.Login)
				assert.NotEmpty(t, d.PasswordHash)
				assert.Equal(t, "John Smith", d.DisplayName)
				assert.Equal(t, "js@sample.com", d.Email)
				assert.EqualValues(t, refinedBody, string(d.Resources))
			}
		}

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, personURL+"/"+thePerson.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err = http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			p := new(model.Person)
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			err = json.Unmarshal(bb, p)
			require.NoError(t, err)

			assert.EqualValues(t, refinedBody, p.Resources)
		}
	})

	t.Run("put resources as stranger", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		theStranger, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			DisplayName:  "Stranger",
			Email:        "s@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"GitHub":"https://github.com/almaz-uno",
			"Telegram messenger":"https://t.me/almaz_develop_bot"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+theStranger.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "insufficient rights", e.Message)
		}
	})

	t.Run("put resources invalid JSON", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{ so funny
			bad bad JSON!!!"GitHub":"https://github.com/almaz-uno",
			Telegram messenger":"https://t.me/almaz_develop_bot"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "invalid format", e["message"])
		}
	})

	t.Run("put resources not authorized", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"GitHub":"https://github.com/almaz-uno",
			"Telegram messenger":"https://t.me/almaz_develop_bot"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID+"/resources", bytes.NewReader([]byte(body)))
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

	t.Run("patch not found", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+pgdao.NewID()+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "insufficient rights", e.Message)
		}
	})
}

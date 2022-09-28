package intest

import (
	"bytes"
	"database/sql"
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

	require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
		IsAdmin: true,
		ID:      creator.ID,
	}))

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.AccessToken)

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.AccessToken)

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.AccessToken)

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+creator.AccessToken)

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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.BasicPersonDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, thePerson.ID, e.ID)
			assert.Equal(t, "0x1234567890abcd", e.EthereumAddress)

			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "0x1234567890abcd", d.EthereumAddress)
			}
		}
	})

	t.Run("patch display_name ok", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"display_name":"Jude Law"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.BasicPersonDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, thePerson.ID, e.ID)
			assert.Equal(t, "Jude Law", e.DisplayName)

			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "Jude Law", d.DisplayName)
			}
		}
	})

	t.Run("patch display_name will not change the original value if new value is an empty string", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "Original Name",
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"display_name":" "
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.BasicPersonDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, thePerson.ID, e.ID)
			assert.Equal(t, "Original Name", e.DisplayName)

			d, err := queries.PersonGet(ctx, thePerson.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, thePerson.ID, d.ID)
				assert.Equal(t, "Original Name", d.DisplayName)
			}
		}
	})

	t.Run("patch invalid person", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
		})
		require.NoError(t, err)

		theStranger, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("1234"),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+theStranger.AccessToken.String)

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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+pgdao.NewID(), bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

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

	t.Run("put resources to new person", func(t *testing.T) {
		thePerson, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        "inhouse",
			Login:        pgdao.NewID(),
			PasswordHash: pgsvc.CreateHashFromPassword("abcd"),
			DisplayName:  "John Smith",
			Email:        "js@sample.com",
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
			IsAdmin: true,
			ID:      thePerson.ID,
		}))

		body := `{
			"GitHub": "https://github.com/almaz-uno",
			 "Telegram messenger": "https://t.me/almaz_develop_bot"
		}`

		refinedBody := strings.NewReplacer("\t", "", "\n", "").Replace(body)

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+thePerson.ID+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
			IsAdmin: true,
			ID:      thePerson.ID,
		}))

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+theStranger.AccessToken.String)

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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
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
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "body is not properly formatted json", e["message"])
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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"ethereum_address":"0x1234567890abcd"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, personURL+"/"+pgdao.NewID()+"/resources", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+thePerson.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "insufficient rights", e.Message)
		}
	})
}

func TestIsAdmin(t *testing.T) {
	require.NoError(t, pgdao.PurgeDB(ctx, db))

	strangerUser, err := pgsvc.NewPerson(db).Add(ctx, &model.Person{
		Login:    "stranger",
		Password: "12345678",
	})
	require.NoError(t, err)

	passengerUser, err := pgsvc.NewPerson(db).Add(ctx, &model.Person{
		Login:    "passenger",
		Password: "12345678",
	})
	require.NoError(t, err)

	adminUser, err := pgsvc.NewPerson(db).Add(ctx, &model.Person{
		Login:    "admin",
		Password: "12345678",
	})
	require.NoError(t, err)

	require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
		IsAdmin: true,
		ID:      adminUser.ID,
	}))

	unauthorized := func(method, url string) func(t *testing.T) {
		return func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString("{}"))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+strangerUser.AccessToken)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := model.BackendError{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.EqualValues(t, "insufficient rights", e.Message)
			}
		}
	}

	authorized := func(method, url string) func(t *testing.T) {
		return func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString("{}"))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+adminUser.AccessToken)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.True(t, int(res.StatusCode/100) == 2, "Invalid result status code '%s' (must be 2xx)", res.Status) {
			}
		}
	}

	rr := strings.NewReplacer("/", ">")

	cases := []struct{ method, url string }{
		{http.MethodPost, "/persons"},
		{http.MethodGet, "/persons"},
		{http.MethodGet, "/persons/" + passengerUser.ID},
	}

	for _, c := range cases {
		t.Run(rr.Replace(c.method+" "+c.url+" unauthorized"), unauthorized(c.method, appURL+c.url))
	}

	for _, c := range cases {
		t.Run(rr.Replace(c.method+" "+c.url+" authorized"), authorized(c.method, appURL+c.url))
	}
}

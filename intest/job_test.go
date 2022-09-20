package intest

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

func TestJob(t *testing.T) {
	faker := faker.New()
	resourceName := "jobs"
	startURL := appURL + "/" + resourceName

	var createdID string

	require.NoError(t, pgdao.PurgeDB(ctx, db))

	createdBy, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "Person1",
		AccessToken: sql.NullString{
			String: pgdao.NewID(),
			Valid:  true,
		},
	})
	require.NoError(t, err)

	require.NoError(t, queries.PersonSetEthereumAddress(ctx, pgdao.PersonSetEthereumAddressParams{
		EthereumAddress: faker.Crypto().EtheriumAddress(),
		ID:              createdBy.ID,
	}))

	personWithoutWallet, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "Person2",
		AccessToken: sql.NullString{
			String: pgdao.NewID(),
			Valid:  true,
		},
	})
	require.NoError(t, err)

	t.Run("get•empty", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Job, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("post•without wallet", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very many words."
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+personWithoutWallet.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Ethereum address is required", e["message"])
		}
	})

	t.Run("post•budget has an invalid format", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very many words.",
			"budget": "",
			"duration": 30
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "invalid format", e["message"])
		}
	})

	t.Run("post•budget negative", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very many words.",
			"budget": -100.2,
			"duration": 30
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Budget must be positive", e["message"])
		}
	})

	t.Run("post•401", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very many words.",
			"budget": 100.2,
			"duration": 30
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
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

	t.Run("post•full", func(t *testing.T) {
		body := `{
			"title":"Create awesome site",
			"description": "There are words here. Very many words.",
			"budget": "100.2",
			"duration": 30
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			createdID = e.ID
			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.Equal(t, "Create awesome site", e.Title)
			assert.Equal(t, "There are words here. Very many words.", e.Description)
			assert.True(t, decimal.RequireFromString("100.2").Equal(e.Budget))
			assert.EqualValues(t, 30, e.Duration)
			assert.NotEmpty(t, e.CreatedAt)
			assert.Equal(t, createdBy.ID, e.CreatedBy)
			assert.NotEmpty(t, e.UpdatedAt)

			d, err := pgdao.New(db).JobGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.Title, d.Title)
				assert.Equal(t, e.Description, d.Description)
				assert.Equal(t, "100.2", d.Budget.String)
				assert.Equal(t, e.Duration, d.Duration.Int32)

				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())

				assert.Equal(t, createdBy.ID, d.CreatedBy)
			}
		}
	})

	t.Run("post•wo-optional", func(t *testing.T) {
		body := `{
			"title":"Create awesome site (wo optional)",
			"description": "There are words here. Very many words. Without optional fields."
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.Equal(t, "Create awesome site (wo optional)", e.Title)
			assert.Equal(t, "There are words here. Very many words. Without optional fields.", e.Description)
			assert.True(t, decimal.Zero.Equal(e.Budget))
			assert.EqualValues(t, 0, e.Duration)
			assert.NotEmpty(t, e.CreatedAt)
			assert.Equal(t, createdBy.ID, e.CreatedBy)
			assert.NotEmpty(t, e.UpdatedAt)

			d, err := pgdao.New(db).JobGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.Title, d.Title)
				assert.Equal(t, e.Description, d.Description)
				assert.False(t, d.Budget.Valid)
				assert.False(t, d.Duration.Valid)

				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())

				assert.Equal(t, createdBy.ID, d.CreatedBy)
			}

		}
	})

	t.Run("post•required-fields", func(t *testing.T) {
		body := `{
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, startURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "Title is required", e.Message)
		}
	})

	t.Run("get", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]model.Job, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			if assert.NotEmpty(t, ee) {
				for _, e := range ee {
					assert.NotEmpty(t, e.ID)
					assert.NotEmpty(t, e.Title)
					assert.NotEmpty(t, e.Description)
					assert.NotEmpty(t, e.CreatedAt)
					assert.Empty(t, e.UpdatedAt)
					assert.NotEmpty(t, e.CreatedBy)
					assert.EqualValues(t, 0, e.ApplicationsCount)
				}
			}
		}
	})

	t.Run("get/:id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL+"/"+createdID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Create awesome site", e.Title)
				assert.Equal(t, "There are words here. Very many words.", e.Description)
				assert.True(t, decimal.RequireFromString("100.2").Equal(e.Budget))
				assert.EqualValues(t, 30, e.Duration)
				assert.NotEmpty(t, e.CreatedAt)
				assert.Equal(t, createdBy.ID, e.CreatedBy)
				assert.NotEmpty(t, e.UpdatedAt)
				assert.Equal(t, uint(0), e.ApplicationsCount)
			}
		}
	})

	t.Run("get/:id•not-found", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL+"/"+"invalid-id", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Entity with specified id not found", e["message"])
		}
	})
}

func TestJobEdit(t *testing.T) {
	ctx := context.Background()

	startURL := appURL + "/jobs"

	require.NoError(t, pgdao.PurgeDB(ctx, db))
	queries := pgdao.New(db)

	createdBy, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Realm: "inhouse",
		Login: "creator",
		AccessToken: sql.NullString{
			String: pgdao.NewID(),
			Valid:  true,
		},
	})
	require.NoError(t, err)

	stranger, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Realm: "inhouse",
		Login: "stranger",
		AccessToken: sql.NullString{
			String: pgdao.NewID(),
			Valid:  true,
		},
	})
	require.NoError(t, err)

	t.Run("put•all fields", func(t *testing.T) {
		theJob, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title before change",
			Description: "Description before change",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			Duration: sql.NullInt32{
				Int32: 24,
				Valid: true,
			},
			CreatedBy: createdBy.ID,
		})
		require.NoError(t, err)

		body := `{
			"title":"Editing title",
			"description": "Editing description. There are words here.",
			"budget": "45.00",
			"duration": 42
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, startURL+"/"+theJob.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Editing title", e.Title)
				assert.Equal(t, "Editing description. There are words here.", e.Description)
				assert.True(t, decimal.RequireFromString("45").Equal(e.Budget))
				assert.EqualValues(t, 42, e.Duration)

				d, err := queries.JobGet(ctx, theJob.ID)
				if assert.NoError(t, err) {
					assert.Equal(t, theJob.ID, d.ID)
					assert.Equal(t, "Editing title", d.Title)
					assert.Equal(t, "Editing description. There are words here.", d.Description)
					assert.Equal(t, "45.00", d.Budget.String)
					assert.EqualValues(t, 42, d.Duration.Int32)
				}
			}
		}
	})

	t.Run("put•budget has an invalid format", func(t *testing.T) {
		theJob, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title before change",
			Description: "Description before change",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			Duration: sql.NullInt32{
				Int32: 24,
				Valid: true,
			},
			CreatedBy: createdBy.ID,
		})
		require.NoError(t, err)

		body := `{
			"title":"Editing title",
			"description": "Editing description. There are words here.",
			"budget": "",
			"duration": 42
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, startURL+"/"+theJob.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Budget has an invalid format", e["message"])
		}
	})

	t.Run("put•budget negative", func(t *testing.T) {
		theJob, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title before change",
			Description: "Description before change",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			Duration: sql.NullInt32{
				Int32: 24,
				Valid: true,
			},
			CreatedBy: createdBy.ID,
		})
		require.NoError(t, err)

		body := `{
			"title":"Editing title",
			"description": "Editing description. There are words here.",
			"budget": "-45.00",
			"duration": 42
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, startURL+"/"+theJob.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Budget must be positive", e["message"])
		}
	})

	t.Run("put•stranger", func(t *testing.T) {
		theJob, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title before change",
			Description: "Description before change",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			Duration: sql.NullInt32{
				Int32: 24,
				Valid: true,
			},
			CreatedBy: createdBy.ID,
		})
		require.NoError(t, err)

		body := `{
			"title":"Editing title",
			"description": "Editing description. There are words here.",
			"budget": "45.0",
			"duration": 42
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, startURL+"/"+theJob.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "entity not found", e.Message)
		}
	})

	t.Run("put•not found", func(t *testing.T) {
		theJob, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title before change",
			Description: "Description before change",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			Duration: sql.NullInt32{
				Int32: 24,
				Valid: true,
			},
			CreatedBy: createdBy.ID,
		})
		require.NoError(t, err)

		body := `{
			"title":"Editing title",
			"description": "Editing description. There are words here.",
			"budget": "45.0",
			"duration": 42
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, startURL+"/abcd"+theJob.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "entity not found", e.Message)
		}
	})
}

func TestBlockedAt(t *testing.T) {
	require.NoError(t, pgdao.PurgeDB(ctx, db))

	theCreator, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        "creator",
		PasswordHash: "123",
		DisplayName:  "The Creator",
		Email:        "creator@sample.com",
		AccessToken: sql.NullString{
			String: "abc",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	theStranger, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        "stranger",
		PasswordHash: "123",
		DisplayName:  "The Stranger",
		Email:        "stranger@sample.com",
		AccessToken: sql.NullString{
			String: "cde",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	theForester, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        "forester",
		PasswordHash: "123",
		DisplayName:  "The Forester",
		Email:        "forester@sample.com",
		AccessToken: sql.NullString{
			String: "kjh",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
		IsAdmin: true,
		ID:      theForester.ID,
	}))

	_, err = queries.JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Number 1",
		Description: "Not blocked",
		CreatedBy:   theCreator.ID,
	})
	require.NoError(t, err)

	j2, err := queries.JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Number 2",
		Description: "Blocked",
		CreatedBy:   theCreator.ID,
	})
	require.NoError(t, err)

	_, err = queries.JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Number 3",
		Description: "Not blocked",
		CreatedBy:   theCreator.ID,
	})
	require.NoError(t, err)

	jj, err := queries.JobsList(ctx)
	require.NoError(t, err)
	require.Len(t, jj, 3)

	j, err := queries.JobGet(ctx, j2.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, j.ID)

	t.Run("block by unauthorized actor", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/jobs/"+j2.ID+"/block", nil)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/jobs/"+j2.ID+"/block", nil)
	require.NoError(t, err)
	req.Header.Set(clog.HeaderXHint, t.Name())
	req.Header.Set(echo.HeaderContentType, "application/json")
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+theForester.AccessToken.String)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
	}

	jj, err = queries.JobsList(ctx)
	require.NoError(t, err)
	require.Len(t, jj, 2)

	_, err = queries.JobGet(ctx, j2.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

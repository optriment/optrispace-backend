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
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

var (
	jobsResourceName = "jobs"
	jobsURL          = appURL + "/" + jobsResourceName
)

func TestCreateJob(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
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

	t.Run("with validation errors", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		t.Run("returns error if body is not a valid JSON", func(t *testing.T) {
			body := `{z}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid JSON", e["message"])
			}
		})

		t.Run("returns error if title is missing", func(t *testing.T) {
			body := `{}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "title is required", e["message"])
			}
		})

		t.Run("returns error if description is missing", func(t *testing.T) {
			body := `{
				"title": "title"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "description is required", e["message"])
			}
		})

		t.Run("returns error if title is an empty string", func(t *testing.T) {
			body := `{
				"title": " ",
				"description": "description"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "title is required", e["message"])
			}
		})

		t.Run("returns error if description is an empty string", func(t *testing.T) {
			body := `{
				"title": "title",
				"description": " "
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "description is required", e["message"])
			}
		})

		t.Run("returns error if budget is not a number", func(t *testing.T) {
			body := `{
				"title": "title",
				"description": "description",
				"budget": "qwe"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid format", e["message"])
			}
		})

		t.Run("returns error if budget is negative", func(t *testing.T) {
			body := `{
				"title": "title",
				"description": "description",
				"budget": -0.1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "budget must be positive", e["message"])
			}
		})

		t.Run("returns error if duration is negative", func(t *testing.T) {
			body := `{
				"title": "title",
				"description": "description",
				"budget": 0.1,
				"duration": -1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "duration must not be negative", e["message"])
			}
		})
	})

	t.Run("returns error if customer does not have ethereum_address", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"title": "title",
			"description": "description",
			"budget": 0.1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "customer does not have wallet", e["message"])
		}
	})

	t.Run("creates job when budget is missing", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		body := `{
			"title": " Title ",
			"description": " Description "
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.JobDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+jobsResourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.NotEmpty(t, e.CreatedAt)
			assert.NotEmpty(t, e.UpdatedAt)
			assert.Equal(t, customer.ID, e.CreatedBy)
			assert.Equal(t, "Title", e.Title)
			assert.Equal(t, "Description", e.Description)
			assert.True(t, decimal.Zero.Equal(e.Budget))
			assert.EqualValues(t, 0, e.Duration)

			d, err := pgdao.New(db).JobGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, "Title", d.Title)
				assert.Equal(t, "Description", d.Description)
				assert.Equal(t, customer.ID, d.CreatedBy)
				assert.False(t, d.Budget.Valid)
				assert.False(t, d.Duration.Valid)
			}
		}
	})

	t.Run("creates job", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		body := `{
			"title": " Title ",
			"description": " Description ",
			"budget": 0.1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.JobDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+jobsResourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.NotEmpty(t, e.CreatedAt)
			assert.NotEmpty(t, e.UpdatedAt)
			assert.Equal(t, customer.ID, e.CreatedBy)
			assert.Equal(t, "Title", e.Title)
			assert.Equal(t, "Description", e.Description)
			assert.True(t, decimal.RequireFromString("0.1").Equal(e.Budget))
			assert.EqualValues(t, 0, e.Duration)

			d, err := pgdao.New(db).JobGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, "Title", d.Title)
				assert.Equal(t, "Description", d.Description)
				assert.Equal(t, customer.ID, d.CreatedBy)
				assert.Equal(t, "0.1", d.Budget.String)
				assert.EqualValues(t, 0, d.Duration.Int32)
			}
		}
	})
}

func TestGetJobs(t *testing.T) {
	t.Run("returns an empty array if there are no jobs", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.JobDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("returns only available jobs ordered by created_at descending", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:          pgdao.NewID(),
			Login:       pgdao.NewID(),
			DisplayName: "Person1",
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: "0x1234567890CUSTOMER1",
		})
		require.NoError(t, err)

		job1, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title1",
			Description: "Description1",
			CreatedBy:   customer1.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:          pgdao.NewID(),
			Login:       pgdao.NewID(),
			DisplayName: "Person2",
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: "0x1234567890CUSTOMER2",
		})
		require.NoError(t, err)

		job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title2",
			Description: "Description2",
			CreatedBy:   customer2.ID,
			Budget: sql.NullString{
				String: "0.2",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		blockedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title3",
			Description: "Description3",
			CreatedBy:   customer2.ID,
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobBlock(ctx, blockedJob.ID)
		require.NoError(t, err)

		suspendedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title4",
			Description: "Description4",
			CreatedBy:   customer2.ID,
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobSuspend(ctx, suspendedJob.ID)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.JobDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			assert.Equal(t, 2, len(ee))

			expectedJob2 := ee[0]
			assert.Equal(t, expectedJob2.ID, job2.ID)
			assert.Equal(t, expectedJob2.Title, "Title2")
			assert.Equal(t, expectedJob2.Description, "Description2")
			assert.Equal(t, expectedJob2.CreatedBy, customer2.ID)
			assert.True(t, decimal.RequireFromString("0.2").Equal(expectedJob2.Budget))
			assert.EqualValues(t, 0, expectedJob2.ApplicationsCount)
			assert.Equal(t, "Person2", expectedJob2.CustomerDisplayName)
			assert.Equal(t, "0x1234567890CUSTOMER2", expectedJob2.CustomerEthereumAddress)

			expectedJob1 := ee[1]
			assert.Equal(t, expectedJob1.ID, job1.ID)
			assert.Equal(t, expectedJob1.Title, "Title1")
			assert.Equal(t, expectedJob1.Description, "Description1")
			assert.Equal(t, expectedJob1.CreatedBy, customer1.ID)
			assert.True(t, decimal.RequireFromString("0.1").Equal(expectedJob1.Budget))
			assert.EqualValues(t, 0, expectedJob1.ApplicationsCount)
			assert.Equal(t, "Person1", expectedJob1.CustomerDisplayName)
			assert.Equal(t, "0x1234567890CUSTOMER1", expectedJob1.CustomerEthereumAddress)
		}
	})
}

func TestGetJob(t *testing.T) {
	t.Run("returns error if job does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobsURL+"/qwerty", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if job is blocked", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		blockedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobBlock(ctx, blockedJob.ID)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobsURL+"/"+blockedJob.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns job if job is available", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:              pgdao.NewID(),
			Login:           pgdao.NewID(),
			DisplayName:     "Person1",
			EthereumAddress: "0x1234567890CUSTOMER1",
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
			Duration:  sql.NullInt32{Int32: 35, Valid: true},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobsURL+"/"+job.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.JobCardDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.Equal(t, job.ID, e.ID)
				assert.Equal(t, "Title", e.Title)
				assert.Equal(t, "Description", e.Description)
				assert.Equal(t, job.CreatedBy, e.CreatedBy)
				assert.Equal(t, job.CreatedAt.UTC(), e.CreatedAt)
				assert.Equal(t, job.UpdatedAt.UTC(), e.UpdatedAt)
				assert.True(t, decimal.RequireFromString("0.1").Equal(e.Budget))
				assert.EqualValues(t, 35, e.Duration)
				assert.EqualValues(t, 0, e.ApplicationsCount)
				assert.Equal(t, "Person1", e.CustomerDisplayName)
				assert.Equal(t, "0x1234567890CUSTOMER1", e.CustomerEthereumAddress)
				assert.False(t, e.IsSuspended)
			}
		}
	})

	t.Run("returns job if job is suspended", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:              pgdao.NewID(),
			Login:           pgdao.NewID(),
			DisplayName:     "Person1",
			EthereumAddress: "0x1234567890CUSTOMER1",
		})
		require.NoError(t, err)

		suspendedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobSuspend(ctx, suspendedJob.ID)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobsURL+"/"+suspendedJob.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.JobCardDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.Equal(t, suspendedJob.ID, e.ID)
				assert.True(t, e.IsSuspended)
			}
		}
	})
}

func TestUpdateJob(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
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

	t.Run("with validation errors", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		t.Run("returns error if body is not a valid JSON", func(t *testing.T) {
			body := `{z}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid JSON", e["message"])
			}
		})

		t.Run("returns error if title is missing", func(t *testing.T) {
			body := `{
				"description": " New description ",
				"budget": 42.1,
				"duration": 1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "title is required", e["message"])
			}
		})

		t.Run("returns error if description is missing", func(t *testing.T) {
			body := `{
				"title": " New title ",
				"budget": 42.1,
				"duration": 1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "description is required", e["message"])
			}
		})

		t.Run("returns error if title is an empty string", func(t *testing.T) {
			body := `{
				"title": " ",
				"description": " New description ",
				"budget": 42.1,
				"duration": 1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "title is required", e["message"])
			}
		})

		t.Run("returns error if description is an empty string", func(t *testing.T) {
			body := `{
				"title": " New title ",
				"description": " ",
				"budget": 42.1,
				"duration": 1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "description is required", e["message"])
			}
		})

		t.Run("returns error if budget is not a number", func(t *testing.T) {
			body := `{
				"title": " New title ",
				"description": " New description ",
				"budget": "qwe",
				"duration": 1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid format", e["message"])
			}
		})

		t.Run("returns error if budget is negative", func(t *testing.T) {
			body := `{
				"title": " New title ",
				"description": " New description ",
				"budget": -0.1,
				"duration": 1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "budget must be positive", e["message"])
			}
		})

		t.Run("returns error if duration is negative", func(t *testing.T) {
			body := `{
				"title": " New title ",
				"description": " New description ",
				"budget": 0.1,
				"duration": -1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "duration must not be negative", e["message"])
			}
		})
	})

	t.Run("returns error if job does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		body := `{
			"title": "Title",
			"description": "Description",
			"budget": 45.0,
			"duration": 42
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/qwerty", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if customer does not have ethereum_address", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
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
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"title": " New title ",
			"description": " New description ",
			"budget": 0.1,
			"duration": 1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/"+job.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "customer does not have wallet", e["message"])
		}
	})

	t.Run("updates job when budget and duration are missing", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
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
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"title": " New title ",
			"description": " New description "
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/"+job.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.JobDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, job.ID, e.ID)
			assert.Equal(t, job.CreatedAt.UTC(), e.CreatedAt)
			assert.Greater(t, e.UpdatedAt, e.CreatedAt)
			assert.Equal(t, customer.ID, e.CreatedBy)
			assert.Equal(t, "New title", e.Title)
			assert.Equal(t, "New description", e.Description)
			assert.True(t, decimal.Zero.Equal(e.Budget))
			assert.EqualValues(t, 0, e.Duration)

			d, err := pgdao.New(db).JobGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, "New title", d.Title)
				assert.Equal(t, "New description", d.Description)
				assert.Equal(t, customer.ID, d.CreatedBy)
				assert.Greater(t, d.UpdatedAt, d.CreatedAt)
				assert.Equal(t, "0", d.Budget.String)
				assert.EqualValues(t, 0, d.Duration.Int32)
			}
		}
	})

	t.Run("updates job", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
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
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"title": " New title ",
			"description": " New description ",
			"budget": 42.1,
			"duration": 35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, jobsURL+"/"+job.ID, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.JobDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, job.ID, e.ID)
			assert.Equal(t, job.CreatedAt.UTC(), e.CreatedAt)
			assert.Greater(t, e.UpdatedAt, e.CreatedAt)
			assert.Equal(t, customer.ID, e.CreatedBy)
			assert.Equal(t, "New title", e.Title)
			assert.Equal(t, "New description", e.Description)
			assert.True(t, decimal.RequireFromString("42.1").Equal(e.Budget))
			assert.EqualValues(t, 35, e.Duration)

			d, err := pgdao.New(db).JobGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, "New title", d.Title)
				assert.Equal(t, "New description", d.Description)
				assert.Equal(t, customer.ID, d.CreatedBy)
				assert.Greater(t, d.UpdatedAt, d.CreatedAt)
				assert.Equal(t, "42.1", d.Budget.String)
				assert.EqualValues(t, 35, d.Duration.Int32)
			}
		}
	})
}

func TestSetJobAsBlocked(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/qwerty/block", nil)
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

	t.Run("returns error if job does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/qwerty/block", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if person is not an admin", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/block", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("returns error if job already blocked", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
			IsAdmin: true,
			ID:      customer.ID,
		}))

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobBlock(ctx, job.ID)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/block", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("blocks job when customer is an admin", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
			IsAdmin: true,
			ID:      customer.ID,
		}))

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/block", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			d, err := pgdao.New(db).JobFind(ctx, job.ID)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, d.BlockedAt)
			}
		}
	})

	t.Run("blocks job when current person is an admin and job belongs to another person", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		admin, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		require.NoError(t, queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
			IsAdmin: true,
			ID:      admin.ID,
		}))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/block", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+admin.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.EqualValues(t, "{}\n", string(bb))

			d, err := pgdao.New(db).JobFind(ctx, job.ID)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, d.BlockedAt)
			}
		}
	})
}

func TestSetJobAsSuspended(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/qwerty/suspend", nil)
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

	t.Run("returns error if job does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/qwerty/suspend", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if a person is not an owner of a job", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		stranger, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/suspend", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("suspends job", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/suspend", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.EqualValues(t, "{}\n", string(bb))

			d, err := pgdao.New(db).JobFind(ctx, job.ID)
			if assert.NoError(t, err) {
				assert.NotEmpty(t, d.SuspendedAt)
			}
		}
	})
}

func TestSetJobAsResumed(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/qwerty/resume", nil)
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

	t.Run("returns error if job does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/qwerty/resume", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if a person is not an owner of a job", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		stranger, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/resume", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("returns error if job is not suspended", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+job.ID+"/resume", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "job is not suspended", e["message"])
		}
	})

	t.Run("resumes job", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		suspendedJob, err := queries.JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			Budget: sql.NullString{
				String: "120.000",
				Valid:  true,
			},
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobSuspend(ctx, suspendedJob.ID)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobsURL+"/"+suspendedJob.ID+"/resume", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.EqualValues(t, "{}\n", string(bb))

			d, err := pgdao.New(db).JobFind(ctx, suspendedJob.ID)
			if assert.NoError(t, err) {
				assert.Empty(t, d.SuspendedAt)
			}
		}
	})
}

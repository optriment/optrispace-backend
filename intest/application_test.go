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
	applicationsResourceName = "applications"
)

func TestCreateApplication(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/jobs/qwerty/applications", bytes.NewReader([]byte(body)))
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
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/applications"

		t.Run("returns error if body is not a valid JSON", func(t *testing.T) {
			body := `{z}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid JSON", e["message"])
			}
		})

		t.Run("returns error if comment is missing", func(t *testing.T) {
			body := `{}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "comment is required", e["message"])
			}
		})

		t.Run("returns error if price is missing", func(t *testing.T) {
			body := `{
				"comment": "comment"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "price is required", e["message"])
			}
		})

		t.Run("returns error if comment is an empty string", func(t *testing.T) {
			body := `{
				"comment": " ",
				"price": 42
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "comment is required", e["message"])
			}
		})

		t.Run("returns error if price is not a number", func(t *testing.T) {
			body := `{
				"comment": "comment",
				"price": "qwe"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid format", e["message"])
			}
		})

		t.Run("returns error if price is negative", func(t *testing.T) {
			body := `{
				"comment": "comment",
				"price": -0.1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "price must be positive", e["message"])
			}
		})

		t.Run("returns error if price is zero", func(t *testing.T) {
			body := `{
				"comment": "comment",
				"price": 0
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "price is required", e["message"])
			}
		})

		t.Run("returns error if applicant does not have ethereum_address", func(t *testing.T) {
			body := `{
				"comment": "comment",
				"price": 0.1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "applicant does not have wallet", e["message"])
			}
		})

		t.Run("returns error if job does not exist", func(t *testing.T) {
			body := `{
				"comment": "comment",
				"price": 0.1
			}`

			notExistentJobApplicationsURL := appURL + "/" + jobsResourceName + "/qwerty/applications"

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, notExistentJobApplicationsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "entity not found", e["message"])
			}
		})
	})

	t.Run("returns error if job is suspended", func(t *testing.T) {
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

		suspendedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		err = pgdao.New(db).JobSuspend(ctx, suspendedJob.ID)
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		suspendedJobURL := appURL + "/" + jobsResourceName + "/" + suspendedJob.ID + "/applications"

		body := `{
			"comment": "comment",
			"price": 0.1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, suspendedJobURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "job does not accept new applications", e["message"])
		}
	})

	t.Run("returns error if customer tries to apply on his own job", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/applications"

		body := `{
			"comment": "comment",
			"price": 0.1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
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

	t.Run("returns error if application already exists", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		applicationParams := pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "comment",
			Price:       "0.1",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		}

		_, err = queries.ApplicationAdd(ctx, applicationParams)
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/applications"

		body := `{
			"comment": "comment",
			"price": 0.1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusConflict, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "application already exists", e["message"])
		}
	})

	t.Run("creates application", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/applications"
		body := `{
			"comment": "Comment",
			"price": 0.1
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, jobApplicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ApplicationDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+applicationsResourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.Equal(t, job.ID, e.JobID)
			assert.Equal(t, applicant.ID, e.ApplicantID)
			assert.Equal(t, "Comment", e.Comment)
			assert.True(t, decimal.RequireFromString("0.1").Equal(e.Price))

			d, err := pgdao.New(db).ApplicationGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, job.ID, d.JobID)
				assert.Equal(t, applicant.ID, d.ApplicantID)
				assert.Equal(t, "Comment", d.Comment)
				assert.Equal(t, e.Price.String(), d.Price)
			}
		}
	})
}

func TestGetMyApplications(t *testing.T) {
	applicationsURL := appURL + "/" + applicationsResourceName

	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, nil)
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

	t.Run("returns an empty array if there are no applications", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ApplicationDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("returns applications belong to a person", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job1, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Job1",
			Description: "Description1",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.101",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Job2",
			Description: "Description2",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.202",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			DisplayName:     "applicant1",
			EthereumAddress: "0xDEADBEEF",
		})
		require.NoError(t, err)

		applicant2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		applicationWithoutContract, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job1.ID,
			ApplicantID: applicant1.ID,
		})
		require.NoError(t, err)

		applicationWithContract, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 2",
			Price:       "0.2",
			JobID:       job2.ID,
			ApplicantID: applicant1.ID,
		})
		require.NoError(t, err)

		_, err = pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 3",
			Price:       "0.3",
			JobID:       job1.ID,
			ApplicantID: applicant2.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Our Contract",
			Description:   "Content",
			Price:         "99.1",
			CustomerID:    customer.ID,
			PerformerID:   applicant1.ID,
			ApplicationID: applicationWithContract.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractCreated,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ApplicationDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Len(t, ee, 2)

			for _, a := range ee {
				switch a.ID {
				case applicationWithoutContract.ID:
					assert.Equal(t, applicant1.ID, a.ApplicantID)
					assert.Equal(t, job1.ID, a.JobID)
					assert.Equal(t, "Job1", a.JobTitle)
					assert.Equal(t, "Description1", a.JobDescription)
					assert.True(t, decimal.RequireFromString("0.101").Equal(a.JobBudget))
					assert.Empty(t, a.ContractID)
					assert.Empty(t, a.ContractStatus)
					assert.Equal(t, "Comment 1", a.Comment)
					assert.True(t, decimal.RequireFromString("42.35").Equal(a.Price))
					assert.Equal(t, applicationWithoutContract.CreatedAt.UTC(), a.CreatedAt)
					assert.Equal(t, "applicant1", a.ApplicantDisplayName)
					assert.Equal(t, "0xDEADBEEF", a.ApplicantEthereumAddress)
				case applicationWithContract.ID:
					assert.Equal(t, applicant1.ID, a.ApplicantID)
					assert.Equal(t, job2.ID, a.JobID)
					assert.Equal(t, "Job2", a.JobTitle)
					assert.Equal(t, "Description2", a.JobDescription)
					assert.True(t, decimal.RequireFromString("0.202").Equal(a.JobBudget))
					assert.Equal(t, contract.ID, a.ContractID)
					assert.Equal(t, "created", a.ContractStatus)
					assert.Equal(t, "Comment 2", a.Comment)
					assert.True(t, decimal.RequireFromString("0.2").Equal(a.Price))
					assert.Equal(t, applicationWithContract.CreatedAt.UTC(), a.CreatedAt)
					assert.Equal(t, "applicant1", a.ApplicantDisplayName)
					assert.Equal(t, "0xDEADBEEF", a.ApplicantEthereumAddress)
				default:
					t.Fail()
				}
			}
		}
	})
}

func TestGetApplication(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		applicationURL := appURL + "/" + applicationsResourceName + "/qwerty"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationURL, nil)
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

	t.Run("returns error if application does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicationURL := appURL + "/" + applicationsResourceName + "/qwerty"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if requested application belongs to another person", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant1.ID,
		})
		require.NoError(t, err)

		applicant2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		applicationURL := appURL + "/" + applicationsResourceName + "/" + application.ID

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("returns application if requested by customer", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:              pgdao.NewID(),
			Login:           pgdao.NewID(),
			DisplayName:     "applicant1",
			EthereumAddress: "0xDEADBEEF",
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		applicationURL := appURL + "/" + applicationsResourceName + "/" + application.ID

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ApplicationDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))

			assert.Equal(t, application.ID, e.ID)
			assert.Equal(t, applicant.ID, e.ApplicantID)
			assert.Equal(t, job.ID, e.JobID)
			assert.Equal(t, "Title", e.JobTitle)
			assert.Equal(t, "Description", e.JobDescription)
			assert.Empty(t, e.ContractID)
			assert.Empty(t, e.ContractStatus)
			assert.Equal(t, "Comment 1", e.Comment)
			assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
			assert.Equal(t, application.CreatedAt.UTC(), e.CreatedAt)
			assert.Equal(t, "applicant1", e.ApplicantDisplayName)
			assert.Equal(t, "0xDEADBEEF", e.ApplicantEthereumAddress)
		}
	})

	t.Run("returns application", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			DisplayName:     "applicant1",
			EthereumAddress: "0xDEADBEEF",
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		applicationURL := appURL + "/" + applicationsResourceName + "/" + application.ID

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ApplicationDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))

			assert.Equal(t, application.ID, e.ID)
			assert.Equal(t, applicant.ID, e.ApplicantID)
			assert.Equal(t, job.ID, e.JobID)
			assert.Equal(t, "Title", e.JobTitle)
			assert.Equal(t, "Description", e.JobDescription)
			assert.Empty(t, e.ContractID)
			assert.Empty(t, e.ContractStatus)
			assert.Equal(t, "Comment 1", e.Comment)
			assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
			assert.Equal(t, application.CreatedAt.UTC(), e.CreatedAt)
			assert.Equal(t, "applicant1", e.ApplicantDisplayName)
			assert.Equal(t, "0xDEADBEEF", e.ApplicantEthereumAddress)
		}
	})
}

func TestGetApplicationsForJob(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/qwerty/" + applicationsResourceName

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationsURL, nil)
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

	t.Run("returns error if requested job belongs to another person", func(t *testing.T) {
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

		anotherCustomer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/" + applicationsResourceName

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+anotherCustomer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
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
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/qwerty/" + applicationsResourceName

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationsURL, nil)
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

	t.Run("returns an empty array if there are no applications", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/" + applicationsResourceName

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ApplicationDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("returns applications for specific job", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:              pgdao.NewID(),
			Login:           pgdao.NewID(),
			DisplayName:     "applicant1",
			EthereumAddress: "0xDEADBEEF",
		})
		require.NoError(t, err)

		applicant2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		applicationWithoutContract, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant1.ID,
		})
		require.NoError(t, err)

		applicationWithContract, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 2",
			Price:       "0.1",
			JobID:       job.ID,
			ApplicantID: applicant2.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Our Contract",
			Description:   "Content",
			Price:         "99.1",
			CustomerID:    customer.ID,
			PerformerID:   applicant2.ID,
			ApplicationID: applicationWithContract.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractCreated,
		})
		require.NoError(t, err)

		anotherJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		_, err = pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 3",
			Price:       "33.1",
			JobID:       anotherJob.ID,
			ApplicantID: applicant1.ID,
		})
		require.NoError(t, err)

		jobApplicationsURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/" + applicationsResourceName

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ApplicationDTO, 2)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			assert.Len(t, ee, 2)

			for _, a := range ee {
				switch a.ID {
				case applicationWithoutContract.ID:
					assert.Equal(t, applicant1.ID, a.ApplicantID)
					assert.Equal(t, job.ID, a.JobID)
					assert.Empty(t, a.ContractID)
					assert.Empty(t, a.ContractStatus)
					assert.Equal(t, "Comment 1", a.Comment)
					assert.True(t, decimal.RequireFromString("42.35").Equal(a.Price))
					assert.Equal(t, applicationWithoutContract.CreatedAt.UTC(), a.CreatedAt)
					assert.Equal(t, "applicant1", a.ApplicantDisplayName)
					assert.Equal(t, "0xDEADBEEF", a.ApplicantEthereumAddress)
				case applicationWithContract.ID:
					assert.Equal(t, applicant2.ID, a.ApplicantID)
					assert.Equal(t, job.ID, a.JobID)
					assert.Equal(t, contract.ID, a.ContractID)
					assert.Equal(t, "created", a.ContractStatus)
					assert.Equal(t, "Comment 2", a.Comment)
					assert.True(t, decimal.RequireFromString("0.1").Equal(a.Price))
					assert.Equal(t, applicationWithContract.CreatedAt.UTC(), a.CreatedAt)
					assert.Equal(t, applicant2.Login, a.ApplicantDisplayName)
					assert.Empty(t, a.ApplicantEthereumAddress)
				default:
					t.Fail()
				}
			}
		}
	})
}

func TestGetApplicationForJob(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		jobApplicationURL := appURL + "/" + jobsResourceName + "/qwerty/application"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationURL, nil)
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

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationURL := appURL + "/" + jobsResourceName + "/qwerty/application"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if application requested by not an applicant", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/application"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationURL, nil)
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

	t.Run("returns empty response if application does not exist", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		jobApplicationURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/application"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.EqualValues(t, "{}\n", string(bb))
		}
	})

	t.Run("returns application", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicationParams := pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "comment",
			Price:       "0.1",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		}

		application, err := queries.ApplicationAdd(ctx, applicationParams)
		require.NoError(t, err)

		jobApplicationURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/application"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ApplicationDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, application.ID, e.ID)
			assert.Equal(t, applicant.ID, e.ApplicantID)
			assert.Equal(t, job.ID, e.JobID)
			assert.Equal(t, "comment", e.Comment)
			assert.True(t, decimal.RequireFromString("0.1").Equal(e.Price))
			assert.Empty(t, e.ContractID)
		}
	})

	t.Run("returns application with contract", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
			Budget: sql.NullString{
				String: "0.1",
				Valid:  true,
			},
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:          pgdao.NewID(),
			Login:       pgdao.NewID(),
			DisplayName: "applicant1",
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: "0xDEADBEEF",
		})
		require.NoError(t, err)

		applicationParams := pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "comment",
			Price:       "0.1",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		}

		application, err := queries.ApplicationAdd(ctx, applicationParams)
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Our Contract",
			Description:   "Content",
			Price:         "99.1",
			CustomerID:    customer.ID,
			PerformerID:   applicant.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractCreated,
		})
		require.NoError(t, err)

		jobApplicationURL := appURL + "/" + jobsResourceName + "/" + job.ID + "/application"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobApplicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ApplicationDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.Equal(t, application.ID, e.ID)
			assert.Equal(t, applicant.ID, e.ApplicantID)
			assert.Equal(t, job.ID, e.JobID)
			assert.Equal(t, contract.ID, e.ContractID)
			assert.Equal(t, "created", e.ContractStatus)
			assert.Equal(t, "comment", e.Comment)
			assert.True(t, decimal.RequireFromString("0.1").Equal(e.Price))
			assert.Equal(t, "applicant1", e.ApplicantDisplayName)
			assert.Equal(t, "0xDEADBEEF", e.ApplicantEthereumAddress)
		}
	})
}

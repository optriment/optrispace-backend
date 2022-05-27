package intest

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

func TestApplication(t *testing.T) {
	resourceName := "applications"

	require.NoError(t, pgdao.PurgeDB(ctx, db))

	stranger, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "stranger",
	})
	require.NoError(t, err)

	createdBy, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "created-by",
	})
	require.NoError(t, err)

	applicant1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "applicant1",
	})
	require.NoError(t, err)

	applicant2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "applicant2",
	})
	require.NoError(t, err)

	applicant3, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "applicant3",
	})
	require.NoError(t, err)

	job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Applications testing",
		Description: "Applications testing description",
		Budget:      sql.NullString{},
		Duration:    sql.NullInt32{},
		CreatedBy:   createdBy.ID,
	})
	require.NoError(t, err)

	jobURL := appURL + "/jobs/" + job.ID
	applicationsURL := jobURL + "/" + resourceName

	// just the job without applications
	t.Run("get•empty-job", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.EqualValues(t, 0, e.ApplicationsCount)
		}
	})

	// just applications for the job without applications (empty list)
	// Should return empty list NOT 404 yet
	t.Run("get•empty-applications", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Application, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	application1, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Do it!",
		Price:       "42.35",
		JobID:       job.ID,
		ApplicantID: applicant1.ID,
	})
	require.NoError(t, err)

	applicationURL := appURL + "/" + resourceName + "/" + application1.ID

	t.Run("get•applications•:id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Do it!", e.Comment)
				assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
				assert.Equal(t, job.ID, e.Job.ID)
				assert.Equal(t, applicant1.ID, e.Applicant.ID)
			}
		}
	})

	// it's required to be authenticated
	t.Run("post•401", func(t *testing.T) {
		body := `{
		}`
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
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

	t.Run("post•comment-required", func(t *testing.T) {
		body := `{
			"price": "123.670000009899232"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "comment: is required", e.Message)
		}
	})

	t.Run("post•price-required", func(t *testing.T) {
		body := `{
			"comment": "Beautiful life!"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "price: is required", e.Message)
		}
	})

	t.Run("job-get•:id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, jobURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Job)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Applications testing", e.Title)
				assert.Equal(t, "Applications testing description", e.Description)
				assert.Equal(t, uint(1), e.ApplicationsCount)
			}
		}
	})

	t.Run("post•applicant2", func(t *testing.T) {
		body := `{
			"comment":"Second one",
			"price": "00003334.77776555"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.NotEmpty(t, e.CreatedAt)
			assert.NotEmpty(t, e.UpdatedAt)
			assert.Equal(t, applicant2.ID, e.Applicant.ID)
			assert.Equal(t, "Second one", e.Comment)
			assert.True(t, decimal.RequireFromString("3334.77776555").Equal(e.Price))
			assert.Equal(t, job.ID, e.Job.ID)
			assert.Nil(t, e.Contract) // there is NO contract yet

			d, err := pgdao.New(db).ApplicationGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())
				assert.Equal(t, e.Comment, d.Comment)

				assert.Equal(t, job.ID, d.JobID)
				assert.Equal(t, applicant2.ID, d.ApplicantID)
			}
		}
	})

	t.Run("post•applicant3", func(t *testing.T) {
		body := `{
			"comment":"رقم ثلاثة",
			"price": "8887.00099990000000"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant3.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.NotEmpty(t, e.CreatedAt)
			assert.NotEmpty(t, e.UpdatedAt)
			assert.Equal(t, applicant3.ID, e.Applicant.ID)
			assert.Equal(t, "رقم ثلاثة", e.Comment)
			assert.True(t, decimal.RequireFromString("8887.0009999").Equal(e.Price))
			assert.Equal(t, job.ID, e.Job.ID)
			assert.Nil(t, e.Contract) // there is NO contract yet

			d, err := pgdao.New(db).ApplicationGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())
				assert.Equal(t, e.Comment, d.Comment)

				assert.Equal(t, job.ID, d.JobID)
				assert.Equal(t, applicant3.ID, d.ApplicantID)
			}
		}
	})

	t.Run("post•repeat-applicant2", func(t *testing.T) {
		body := `{
			"comment":"Repeat",
			"price": "00003334.77776555"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusConflict, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "application already exists", e.Message)
		}
	})

	t.Run("get•:job_id•applications•401", func(t *testing.T) {
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

	contract1, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "99.1",
		Duration:      sql.NullInt32{Int32: 42, Valid: true},
		CustomerID:    createdBy.ID,
		PerformerID:   applicant1.ID,
		ApplicationID: application1.ID,
		CreatedBy:     createdBy.ID,
	})
	require.NoError(t, err)
	_ = contract1

	t.Run("get•:job_id•applications•by-stranger", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Application, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("get•:job_id•applications•by-author", func(t *testing.T) {
		body := ``

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+createdBy.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Application, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			if assert.Len(t, ee, 3) {
				for _, a := range ee {
					assert.NotEmpty(t, a.CreatedAt)
					assert.NotEmpty(t, a.UpdatedAt)
					assert.NotEmpty(t, a.Applicant.ID)
					assert.NotEmpty(t, a.Comment)
					assert.NotEmpty(t, a.Price)
					assert.NotEmpty(t, a.Job)

					if a.ID == application1.ID {
						assert.NotNil(t, a.Contract)
						assert.Equal(t, contract1.ID, a.Contract.ID)
						assert.Equal(t, "created", a.Contract.Status)
						assert.True(t, decimal.RequireFromString("99.1").Equal(a.Contract.Price))
					} else {
						assert.Nil(t, a.Contract)
					}
				}
			}
		}
	})

	t.Run("get•:job_id•applications•by-applicant2", func(t *testing.T) {
		body := ``

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Application, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			if assert.Len(t, ee, 1) {
				for _, a := range ee {
					assert.NotEmpty(t, a.CreatedAt)
					assert.NotEmpty(t, a.UpdatedAt)
					assert.NotEmpty(t, a.Applicant.ID)
					assert.NotEmpty(t, a.Comment)
					assert.NotEmpty(t, a.Price)
					assert.NotEmpty(t, a.Job)
					assert.Nil(t, a.Contract)
				}
			}
		}
	})

	t.Run("get•my applications", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL+"/"+resourceName+"/my", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Application, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			if assert.Len(t, ee, 1) {
				for _, a := range ee {
					assert.NotEmpty(t, a.CreatedAt)
					assert.NotEmpty(t, a.UpdatedAt)
					assert.NotEmpty(t, a.Applicant.ID)
					assert.NotEmpty(t, a.Comment)
					assert.NotEmpty(t, a.Price)

					assert.NotNil(t, a.Job)
					assert.Equal(t, job.ID, a.Job.ID)
					assert.Equal(t, a.Job.Title, "Applications testing")
					assert.Equal(t, a.Job.Description, "Applications testing description")

					assert.NotNil(t, a.Contract)
					assert.Equal(t, contract1.ID, a.Contract.ID)
					assert.Equal(t, "created", a.Contract.Status)
					assert.True(t, decimal.RequireFromString("99.1").Equal(a.Contract.Price))
				}
			}
		}
	})
}

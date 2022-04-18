package intest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ryboe/q"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/web"
)

func TestApplication(t *testing.T) {
	resourceName := "applications"

	require.NoError(t, pgdao.New(db).ApplicationsPurge(bgctx))

	require.NoError(t, pgdao.New(db).JobsPurge(bgctx))
	createdBy, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:      pgdao.NewID(),
		Address: pgdao.NewID() + pgdao.NewID(),
	})
	require.NoError(t, err)

	applicant1, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:      pgdao.NewID(),
		Address: pgdao.NewID() + pgdao.NewID(),
	})
	require.NoError(t, err)

	applicant2, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:      pgdao.NewID(),
		Address: pgdao.NewID() + pgdao.NewID(),
	})
	require.NoError(t, err)

	applicant3, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:      pgdao.NewID(),
		Address: pgdao.NewID() + pgdao.NewID(),
	})
	require.NoError(t, err)

	job, err := pgdao.New(db).JobAdd(bgctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Applications testing",
		Description: "Applications testing description",
		Budget:      sql.NullString{},
		Duration:    sql.NullInt32{},
		CreatedBy:   createdBy.ID,
	})
	require.NoError(t, err)

	job2, err := pgdao.New(db).JobAdd(bgctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Applications testing",
		Description: "Applications testing description",
		Budget:      sql.NullString{},
		Duration:    sql.NullInt32{},
		CreatedBy:   createdBy.ID,
	})
	require.NoError(t, err)

	applID := make([]string, 3)

	_ = job2

	jobURL := appURL + "/jobs/" + job.ID
	applicationsURL := jobURL + "/" + resourceName

	// just the job without applications
	t.Run("get•empty-job", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, jobURL, nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
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
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, applicationsURL, nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)
		q.Q(applicationsURL)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Application, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	// it's required to be authenticated
	t.Run("post•401", func(t *testing.T) {
		body := `{
		}`
		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("post•applicant1", func(t *testing.T) {
		body := `{
			"comment":"I will make this easy!",
			"price": "123.670000009899232"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)

		n := time.Now()

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			applID[0] = e.ID

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.WithinDuration(t, n, e.CreatedAt, time.Since(n))
			assert.WithinDuration(t, n, e.UpdatedAt, time.Since(n))
			assert.Equal(t, applicant1.ID, e.Applicant.ID)
			assert.Equal(t, "I will make this easy!", e.Comment)
			assert.True(t, decimal.RequireFromString("123.670000009899232").Equal(e.Price))
			assert.Equal(t, job.ID, e.Job.ID)
			assert.Nil(t, e.Contract) // there is NO contract yet

			d, err := pgdao.New(db).ApplicationGet(bgctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())
				assert.Equal(t, e.Comment, d.Comment)

				assert.Equal(t, job.ID, d.JobID)
				assert.Equal(t, applicant1.ID, d.ApplicantID)
			}
		}
	})

	t.Run("post•comment-required", func(t *testing.T) {
		body := `{
			"price": "123.670000009899232"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Regexp(t, "^Error while processing request:.*$", e["message"])
		}
	})

	t.Run("post•price-required", func(t *testing.T) {
		body := `{
			"comment": "Beautiful life!"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Regexp(t, "^Error while processing request:.*$", e["message"])
		}
	})

	t.Run("job-get/:id", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, jobURL, nil)
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())

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

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

		n := time.Now()

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			applID[1] = e.ID

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.WithinDuration(t, n, e.CreatedAt, time.Since(n))
			assert.WithinDuration(t, n, e.UpdatedAt, time.Since(n))
			assert.Equal(t, applicant2.ID, e.Applicant.ID)
			assert.Equal(t, "Second one", e.Comment)
			assert.True(t, decimal.RequireFromString("3334.77776555").Equal(e.Price))
			assert.Equal(t, job.ID, e.Job.ID)
			assert.Nil(t, e.Contract) // there is NO contract yet

			d, err := pgdao.New(db).ApplicationGet(bgctx, e.ID)
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

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant3.ID)

		n := time.Now()

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			applID[2] = e.ID

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.WithinDuration(t, n, e.CreatedAt, time.Since(n))
			assert.WithinDuration(t, n, e.UpdatedAt, time.Since(n))
			assert.Equal(t, applicant3.ID, e.Applicant.ID)
			assert.Equal(t, "رقم ثلاثة", e.Comment)
			assert.True(t, decimal.RequireFromString("8887.0009999").Equal(e.Price))
			assert.Equal(t, job.ID, e.Job.ID)
			assert.Nil(t, e.Contract) // there is NO contract yet

			d, err := pgdao.New(db).ApplicationGet(bgctx, e.ID)
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

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusConflict, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Regexp(t, `^Duplication: unable to create application &{Comment:Repeat Price:3334.77776555}: Key \(job_id, applicant_id\)=\(\w+, \w+\) already exists\.: duplication$`, e["message"])
		}
	})

	t.Run("get/:job_id/applications•401", func(t *testing.T) {
		body := ``

		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		// req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("get/:job_id/applications•successfully", func(t *testing.T) {
		body := ``

		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, applicationsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(web.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant2.ID)

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
					assert.Nil(t, a.Contract)
				}
			}
		}
	})
}

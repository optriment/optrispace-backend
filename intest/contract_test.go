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

func TestContract(t *testing.T) {
	resourceName := "contracts"
	contractsURL := appURL + "/" + resourceName

	require.NoError(t, pgdao.PurgeDB(bgctx, db))

	t.Run("get /contracts should be protected for unauthorized request", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL, nil)
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

	customer1, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "customer1",
	})
	require.NoError(t, err)

	t.Run("get /contracts returns empty array", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Contract, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	performer1, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "performer1",
	})
	require.NoError(t, err)

	job, err := pgdao.New(db).JobAdd(bgctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Contracts testing",
		Description: "Contracts testing description",
		CreatedBy:   customer1.ID,
	})
	require.NoError(t, err)

	application1, err := pgdao.New(db).ApplicationAdd(bgctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Do it!",
		JobID:       job.ID,
		Price:       "42.35",
		ApplicantID: performer1.ID,
	})
	require.NoError(t, err)

	contract1, err := pgdao.New(db).ContractAdd(bgctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Do it!",
		Description:   "Descriptive message",
		Price:         "42.35",
		Duration:      sql.NullInt32{Int32: 35, Valid: true},
		CustomerID:    customer1.ID,
		PerformerID:   performer1.ID,
		ApplicationID: application1.ID,
		CreatedBy:     customer1.ID,
	})
	require.NoError(t, err)

	customer2, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "customer2",
	})
	require.NoError(t, err)

	job2, err := pgdao.New(db).JobAdd(bgctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Title",
		Description: "Description",
		CreatedBy:   customer2.ID,
	})
	require.NoError(t, err)

	performer2, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "performer2",
	})
	require.NoError(t, err)

	application2, err := pgdao.New(db).ApplicationAdd(bgctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "I can do it",
		JobID:       job2.ID,
		Price:       "7.99",
		ApplicantID: performer2.ID,
	})
	require.NoError(t, err)

	contract2, err := pgdao.New(db).ContractAdd(bgctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Do it again!",
		Description:   "Descriptive message 2",
		Price:         "35.42",
		CustomerID:    customer2.ID,
		PerformerID:   performer2.ID,
		ApplicationID: application2.ID,
		CreatedBy:     customer2.ID,
	})
	require.NoError(t, err)

	_ = contract2

	t.Run("get /contracts returns only owned contracts", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.Contract, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			// assert.Empty(t, ee)
			assert.Equal(t, 1, len(ee))

			result := ee[0]
			assert.Equal(t, contract1.ID, result.ID)
		}
	})

	t.Run("get /contracts/:id should be success for authorized customer", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Contract)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Do it!", e.Title)
				assert.Equal(t, "Descriptive message", e.Description)
				assert.Equal(t, customer1.ID, e.Customer.ID)
				assert.Equal(t, performer1.ID, e.Performer.ID)
				assert.Equal(t, customer1.ID, e.CreatedBy)
				assert.Equal(t, application1.ID, e.Application.ID)
				assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
				assert.EqualValues(t, 35, e.Duration)
			}
		}
	})

	t.Run("get /contracts/:id should be success for authorized performer", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Contract)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, "Do it!", e.Title)
				assert.Equal(t, "Descriptive message", e.Description)
				assert.Equal(t, customer1.ID, e.Customer.ID)
				assert.Equal(t, performer1.ID, e.Performer.ID)
				assert.Equal(t, customer1.ID, e.CreatedBy)
				assert.Equal(t, application1.ID, e.Application.ID)
				assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
				assert.EqualValues(t, 35, e.Duration)
			}
		}
	})

	t.Run("get /contracts/:id should not be found for unauthorized customer", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("get /contracts/:id should not be found for unauthorized performer", func(t *testing.T) {
		req, err := http.NewRequestWithContext(bgctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	// it's required to be authenticated
	t.Run("post•401", func(t *testing.T) {
		body := `{}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

	performer3, err := pgdao.New(db).PersonAdd(bgctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "performer3",
	})
	require.NoError(t, err)

	application3, err := pgdao.New(db).ApplicationAdd(bgctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Do it!",
		JobID:       job.ID,
		Price:       "42.35",
		ApplicantID: performer3.ID,
	})
	require.NoError(t, err)

	t.Run("post•contract", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"duration": 42,
			"performer_id": "` + performer3.ID + `",
			"application_id": "` + application3.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Contract)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+resourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.NotEmpty(t, e.CreatedAt)
			assert.NotEmpty(t, e.UpdatedAt)
			assert.Equal(t, customer1.ID, e.Customer.ID)
			assert.Equal(t, performer3.ID, e.Performer.ID)
			assert.Equal(t, application3.ID, e.Application.ID)
			assert.Equal(t, "I will make this easy!", e.Title)
			assert.Equal(t, "I believe in you!", e.Description)
			assert.EqualValues(t, 42, e.Duration)
			assert.True(t, decimal.RequireFromString("123.670000009899232").Equal(e.Price))
		}
	})

	t.Run("post•application_id required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Contains(t, e["message"], "application_id required")
		}
	})

	t.Run("post•performer_id required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Contains(t, e["message"], "performer_id required")
		}
	})

	t.Run("post•title required", func(t *testing.T) {
		body := `{
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Contains(t, e["message"], "title required")
		}
	})

	t.Run("post•description required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Contains(t, e["message"], "description required")
		}
	})

	t.Run("post•price required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "0",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))

			assert.Contains(t, e["message"], "price required")
		}
	})

	t.Run("post•price negative", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "-1.0",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(bgctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))

			assert.Contains(t, e["message"], "price must be positive")
		}
	})

	// TODO: When the customer creates a contract for the existing application
}
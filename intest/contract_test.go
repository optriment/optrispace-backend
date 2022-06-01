package intest

import (
	"bytes"
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

func TestContract(t *testing.T) {
	resourceName := "contracts"
	contractsURL := appURL + "/" + resourceName

	require.NoError(t, pgdao.PurgeDB(ctx, db))

	t.Run("get /contracts should be protected for unauthorized request", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL, nil)
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

	customer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "customer1",
	})
	require.NoError(t, err)

	t.Run("get /contracts returns empty array", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL, nil)
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

	performer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "performer1",
	})
	require.NoError(t, err)

	job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Contracts testing",
		Description: "Contracts testing description",
		CreatedBy:   customer1.ID,
	})
	require.NoError(t, err)

	application1, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Do it!",
		JobID:       job.ID,
		Price:       "42.35",
		ApplicantID: performer1.ID,
	})
	require.NoError(t, err)

	contract1, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
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

	customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "customer2",
	})
	require.NoError(t, err)

	job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Title",
		Description: "Description",
		CreatedBy:   customer2.ID,
	})
	require.NoError(t, err)

	performer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "performer2",
	})
	require.NoError(t, err)

	application2, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "I can do it",
		JobID:       job2.ID,
		Price:       "7.99",
		ApplicantID: performer2.ID,
	})
	require.NoError(t, err)

	contract2, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
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
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL, nil)
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
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
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
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
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
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("get /contracts/:id should not be found for unauthorized performer", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract1.ID, nil)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

	performer3, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:    pgdao.NewID(),
		Login: "performer3",
	})
	require.NoError(t, err)

	application3, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
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
			"application_id": "` + application3.ID + `",
			"customer_address":"0x1234567890customer"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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
			assert.Equal(t, "0x1234567890customer", e.CustomerAddress)

			d, err := pgdao.New(db).ContractGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.Customer.ID, d.CustomerID)
				assert.Equal(t, e.Performer.ID, d.PerformerID)
				assert.Equal(t, e.Application.ID, d.ApplicationID)
				assert.Equal(t, e.Title, d.Title)
				assert.Equal(t, e.Description, d.Description)
				assert.Equal(t, e.Price.String(), d.Price)
				assert.EqualValues(t, e.Duration, d.Duration.Int32)
				assert.Equal(t, e.Status, d.Status)
				assert.Equal(t, e.CreatedBy, d.CreatedBy)
				assert.Equal(t, e.CreatedAt, d.CreatedAt.UTC())
				assert.Equal(t, e.UpdatedAt, d.UpdatedAt.UTC())
				assert.Equal(t, e.CustomerAddress, d.CustomerAddress)
				assert.Equal(t, e.PerformerAddress, d.PerformerAddress)
				assert.Equal(t, e.CustomerAddress, d.CustomerAddress)
			}
		}
	})

	t.Run("post•application_id required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "application_id: is required", e.Message)
		}
	})

	t.Run("post•customer_address required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application3.ID + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "customer_address: is required", e.Message)
		}
	})

	t.Run("post•performer_id required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "performer_id: is required", e.Message)
		}
	})

	t.Run("post•title required", func(t *testing.T) {
		body := `{
			"description":"I believe in you!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "title: is required", e.Message)
		}
	})

	t.Run("post•description required", func(t *testing.T) {
		body := `{
			"title":"I will make this easy!",
			"price": "123.670000009899232",
			"performer_id": "` + performer1.ID + `",
			"application_id": "` + application1.ID + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "description: is required", e.Message)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "price: is required", e.Message)
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

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.ID)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "price: must be positive", e.Message)
		}
	})

	// TODO: When the customer creates a contract for the existing application
}

func TestContractStatuses(t *testing.T) {
	queries := pgdao.New(db)

	require.NoError(t, pgdao.PurgeDB(ctx, db))

	stranger, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        "stranger",
		PasswordHash: "123456",
		DisplayName:  "stranger",
		Email:        "stranger@sample.com",
	})
	require.NoError(t, err)

	customer, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        "customer",
		PasswordHash: "123456",
		DisplayName:  "customer",
		Email:        "customer@sample.com",
	})
	require.NoError(t, err)

	performer, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        "performer",
		PasswordHash: "123456",
		DisplayName:  "performer",
		Email:        "performer@sample.com",
	})
	require.NoError(t, err)

	job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Some job",
		Description: faker.New().Letter(),
		Budget: sql.NullString{
			String: "20.00",
			Valid:  true,
		},
		Duration:  sql.NullInt32{},
		CreatedBy: customer.ID,
	})
	require.NoError(t, err)

	application, err := queries.ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     faker.New().Letter(),
		Price:       "18.9",
		JobID:       job.ID,
		ApplicantID: performer.ID,
	})
	require.NoError(t, err)

	contract, err := queries.ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		CustomerID:    customer.ID,
		PerformerID:   performer.ID,
		ApplicationID: application.ID,
		Title:         "Some awesome job",
		Description:   faker.New().Letter(),
		Price:         "19.0",
		Duration: sql.NullInt32{
			Int32: 9,
			Valid: true,
		},
		CreatedBy: customer.ID,
	})

	resourceName := "contracts"
	contractsURL := appURL + "/" + resourceName
	theContractURL := contractsURL + "/" + contract.ID

	notFoundTest := func(action, body string) func(t *testing.T) {
		return func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/non-existing-id/"+action, bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.ID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := model.BackendError{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.EqualValues(t, "entity not found", e.Message)
			}
		}
	}

	badRequest := func(action, body string) func(t *testing.T) {
		return func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/non-existing-id/"+action, bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.ID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))

				assert.Contains(t, e["message"], "Syntax error")
			}
		}
	}

	strangerTest := func(action, startStatus, body string) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := queries.ContractPatch(ctx, pgdao.ContractPatchParams{
				StatusChange: true,
				Status:       startStatus,
				ID:           contract.ID,
			})
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, theContractURL+"/"+action, bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.ID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := model.BackendError{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.EqualValues(t, "entity not found", e.Message)
			}
		}
	}

	invalidActorTest := func(action, startStatus, actorID, body string) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := queries.ContractPatch(ctx, pgdao.ContractPatchParams{
				StatusChange: true,
				Status:       startStatus,
				ID:           contract.ID,
			})
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, theContractURL+"/"+action, bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+actorID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := model.BackendError{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.EqualValues(t, "insufficient rights", e.Message)
			}
		}
	}

	okTest := func(action, startStatus, targetStatus, actorID, body string, verifier func(t *testing.T, c *pgdao.Contract)) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := queries.ContractPatch(ctx, pgdao.ContractPatchParams{
				StatusChange: true,
				Status:       startStatus,
				ID:           contract.ID,
			})
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, theContractURL+"/"+action, bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+actorID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				c, err := queries.ContractGet(ctx, contract.ID)
				if assert.NoError(t, err) {
					assert.Equal(t, targetStatus, c.Status)
					if verifier != nil {
						verifier(t, &c)
					}
				}
			}
		}
	}

	invalidSourceStatusTest := func(action, startStatus, actorID, body string) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := queries.ContractPatch(ctx, pgdao.ContractPatchParams{
				StatusChange: true,
				Status:       startStatus,
				ID:           contract.ID,
			})
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, theContractURL+"/"+action, bytes.NewBufferString(body))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+actorID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := model.BackendError{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.EqualValues(t, "inappropriate action", e.Message)
			}
		}
	}

	missedField := func(action, startStatus, actorID, fieldName string) func(t *testing.T) {
		return func(t *testing.T) {
			_, err := queries.ContractPatch(ctx, pgdao.ContractPatchParams{
				StatusChange: true,
				Status:       startStatus,
				ID:           contract.ID,
			})
			require.NoError(t, err)

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, theContractURL+"/"+action, nil)
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+actorID)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := model.BackendError{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.EqualValues(t, fieldName+": is required", e.Message)
			}
		}
	}

	action := "accept"
	t.Run(action, func(t *testing.T) {
		t.Run("not-found", notFoundTest(action, `{"performer_address":"0x123456abcd"}`))
		t.Run("bad-request", badRequest(action, `aaasdddddsssssd`))
		t.Run("stranger", strangerTest(action, model.ContractCreated, `{"performer_address":"0x123456abcd"}`))
		t.Run("customer", invalidActorTest(action, model.ContractCreated, customer.ID, `{"performer_address":"0x123456abcd"}`))
		t.Run("missed performer_address", missedField(action, model.ContractCreated, customer.ID, "performer_address"))
		t.Run("performer", okTest(action, model.ContractCreated, model.ContractAccepted, performer.ID, `{"performer_address":"0x123456abcd"}`, func(t *testing.T, c *pgdao.Contract) {
			assert.Equal(t, "0x123456abcd", c.PerformerAddress)
		}))
		for _, st := range []string{
			// model.ContractCreated,
			model.ContractAccepted,
			model.ContractDeployed,
			model.ContractSent,
			model.ContractApproved,
			model.ContractCompleted,
		} {
			t.Run("status "+st, invalidSourceStatusTest(action, st, performer.ID, `{"performer_address":"0x123456abcd"}`))
		}
	})

	action = "deploy"
	t.Run(action, func(t *testing.T) {
		t.Run("not-found", notFoundTest(action, `{"contract_address":"0x0987654dsa"}`))
		t.Run("stranger", strangerTest(action, model.ContractAccepted, `{"contract_address":"0x0987654dsa"}`))
		t.Run("customer", okTest(action, model.ContractAccepted, model.ContractDeployed, customer.ID, `{"contract_address":"0x0987654dsa"}`, func(t *testing.T, c *pgdao.Contract) {
			assert.Equal(t, "0x0987654dsa", c.ContractAddress)
		}))
		t.Run("performer", invalidActorTest(action, model.ContractAccepted, performer.ID, `{"contract_address":"0x0987654dsa"}`))
		t.Run("missed contract_address", missedField(action, model.ContractAccepted, customer.ID, "contract_address"))
		for _, st := range []string{
			model.ContractCreated,
			// model.ContractAccepted,
			model.ContractDeployed,
			model.ContractSent,
			model.ContractApproved,
			model.ContractCompleted,
		} {
			t.Run("status "+st, invalidSourceStatusTest(action, st, customer.ID, `{"contract_address":"0x0987654dsa"}`))
		}
	})

	action = "send"
	t.Run(action, func(t *testing.T) {
		t.Run("not-found", notFoundTest(action, ""))
		t.Run("stranger", strangerTest(action, model.ContractDeployed, ""))
		t.Run("customer", invalidActorTest(action, model.ContractDeployed, customer.ID, ""))
		t.Run("performer", okTest(action, model.ContractDeployed, model.ContractSent, performer.ID, "", nil))
		for _, st := range []string{
			model.ContractCreated,
			model.ContractAccepted,
			// model.ContractDeployed,
			model.ContractSent,
			model.ContractApproved,
			model.ContractCompleted,
		} {
			t.Run("status "+st, invalidSourceStatusTest(action, st, performer.ID, ""))
		}
	})

	action = "approve"
	t.Run(action, func(t *testing.T) {
		t.Run("not-found", notFoundTest(action, ""))
		t.Run("stranger", strangerTest(action, model.ContractSent, ""))
		t.Run("customer", okTest(action, model.ContractSent, model.ContractApproved, customer.ID, "", nil))
		t.Run("performer", invalidActorTest(action, model.ContractSent, performer.ID, ""))
		for _, st := range []string{
			model.ContractCreated,
			model.ContractAccepted,
			model.ContractDeployed,
			// model.ContractSent,
			model.ContractApproved,
			model.ContractCompleted,
		} {
			t.Run("status "+st, invalidSourceStatusTest(action, st, customer.ID, ""))
		}
	})

	action = "complete"
	t.Run(action, func(t *testing.T) {
		t.Run("not-found", notFoundTest(action, ""))
		t.Run("stranger", strangerTest(action, model.ContractApproved, ""))
		t.Run("customer", okTest(action, model.ContractApproved, model.ContractCompleted, customer.ID, "", nil))
		t.Run("performer", invalidActorTest(action, model.ContractApproved, performer.ID, ""))
		for _, st := range []string{
			model.ContractCreated,
			model.ContractAccepted,
			model.ContractDeployed,
			model.ContractSent,
			// model.ContractApproved,
			model.ContractCompleted,
		} {
			t.Run("status "+st, invalidSourceStatusTest(action, st, customer.ID, ""))
		}
	})
}

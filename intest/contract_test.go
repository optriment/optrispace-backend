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

var (
	contractsResourceName = "contracts"
	contractsURL          = appURL + "/" + contractsResourceName
)

const (
	validBlockchainAddress   = "0x8Ca2702c5bcc50D79d9a059D58607028aa36Aa6c"
	fundedContractAddress    = "0xaB8722B889D231d62c9eB35Eb1b557926F3B3289"
	notFundedContractAddress = "0x9Ca2702c5bcc51D79d9a059D58607028aa36DD67"
)

func TestCreateContract(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

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

		t.Run("returns error if body is not a valid JSON", func(t *testing.T) {
			body := `{z}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

		t.Run("returns error if application_id is missing", func(t *testing.T) {
			body := `{}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "application_id is required", e["message"])
			}
		})

		t.Run("returns error if title is missing", func(t *testing.T) {
			body := `{
				"application_id": "qwerty"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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
				"application_id": "qwerty",
				"title": "title"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

		t.Run("returns error if price is missing", func(t *testing.T) {
			body := `{
				"application_id": "qwerty",
				"title": "title",
				"description": "description"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "price is required", e["message"])
			}
		})

		t.Run("returns error if application_id is an empty string", func(t *testing.T) {
			body := `{
				"application_id": " ",
				"title": " ",
				"description": " ",
				"price": "0.1"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "application_id is required", e["message"])
			}
		})

		t.Run("returns error if title is an empty string", func(t *testing.T) {
			body := `{
				"application_id": "qwerty",
				"title": " ",
				"description": "description",
				"price": 0.1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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
				"application_id": "qwerty",
				"title": "title",
				"description": " ",
				"price": 0.1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

		t.Run("returns error if price is not a number", func(t *testing.T) {
			body := `{
				"application_id": " ",
				"title": " ",
				"description": " ",
				"price": "qwe"
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"?application_id=zzzzzz", bytes.NewReader([]byte(body)))
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

		t.Run("returns error if price is zero", func(t *testing.T) {
			body := `{
				"application_id": "qwerty",
				"title": "title",
				"description": "description",
				"price": 0
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "price is required", e["message"])
			}
		})

		t.Run("returns error if price is negative", func(t *testing.T) {
			body := `{
				"application_id": "qwerty",
				"title": "title",
				"description": "description",
				"price": -0.1
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "price must be positive", e["message"])
			}
		})

		t.Run("returns error if application does not exist", func(t *testing.T) {
			body := `{
				"application_id": "GgdQfiJGBNfiXQq4DCyPAH",
				"title": "title",
				"description": "description",
				"price": 42.35
			}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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
	})

	t.Run("returns error if application_id does not belong to job owned by customer", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
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

		body := `{
			"application_id": "` + application.ID + `",
			"title": "title",
			"description": "description",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		body := `{
			"application_id": "` + application.ID + `",
			"title": "title",
			"description": "description",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
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

	t.Run("returns error if performer does not have ethereum_address", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		body := `{
			"application_id": "` + application.ID + `",
			"title": "title",
			"description": "description",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "performer does not have wallet", e["message"])
		}
	})

	t.Run("returns error if customer and performer are the same person", func(t *testing.T) {
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

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"application_id": "` + application.ID + `",
			"title": "title",
			"description": "description",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "inappropriate action", e["message"])
		}
	})

	t.Run("returns error if customer and performer have the same ethereum_address", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:              pgdao.NewID(),
			Login:           pgdao.NewID(),
			EthereumAddress: customer.EthereumAddress,
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		body := `{
			"application_id": "` + application.ID + `",
			"title": "title",
			"description": "description",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "customer and performer addresses cannot be the same", e["message"])
		}
	})

	t.Run("returns error if contract between customer and performer already exists", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:              pgdao.NewID(),
			Login:           pgdao.NewID(),
			EthereumAddress: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"application_id": "` + application.ID + `",
			"title": "title",
			"description": "description",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusConflict, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))

			assert.Equal(t, "contract already exists", e["message"])
		}
	})

	t.Run("creates contract", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: "0x1234567890CUSTOMER",
		})
		require.NoError(t, err)

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: "0x1234567890PERFORMER",
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		body := `{
			"application_id": "` + application.ID + `",
			"title": " Title ",
			"description": " Description ",
			"price": 42.35
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL, bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ContractDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			assert.True(t, strings.HasPrefix(res.Header.Get(echo.HeaderLocation), "/"+contractsResourceName+"/"+e.ID))

			assert.NotEmpty(t, e.ID)
			assert.NotEmpty(t, e.CreatedAt)
			assert.NotEmpty(t, e.UpdatedAt)
			assert.Equal(t, e.CustomerID, customer.ID)
			assert.Equal(t, e.PerformerID, performer.ID)
			assert.Equal(t, e.ApplicationID, application.ID)
			assert.Equal(t, e.Title, "Title")
			assert.Equal(t, e.Description, "Description")
			assert.EqualValues(t, 0, e.Duration)
			assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
			assert.Equal(t, e.ContractAddress, "")
			assert.Equal(t, e.CustomerAddress, "0x1234567890customer")
			assert.Equal(t, e.PerformerAddress, "0x1234567890performer")
			assert.Equal(t, e.Status, "created")

			d, err := pgdao.New(db).ContractGet(ctx, e.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, e.ID, d.ID)
				assert.Equal(t, e.CustomerID, d.CustomerID)
				assert.Equal(t, e.PerformerID, d.PerformerID)
				assert.Equal(t, e.ApplicationID, d.ApplicationID)
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
}

func TestGetContracts(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

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

	t.Run("returns an empty array if there are no contracts for the person", func(t *testing.T) {
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ContractDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("returns only owned contracts for the person", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		performer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			CustomerID:    customer1.ID,
			PerformerID:   performer1.ID,
			ApplicationID: application1.ID,
			CreatedBy:     customer1.ID,
			Status:        model.ContractCreated,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer1.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ContractDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			assert.Equal(t, 1, len(ee))

			result := ee[0]
			assert.Equal(t, result.ID, contract1.ID)
			assert.Equal(t, result.Status, model.ContractCreated)
			assert.Equal(t, result.Title, "Do it!")
			assert.Equal(t, result.Description, "Descriptive message")
			assert.Equal(t, result.CustomerID, customer1.ID)
			assert.Equal(t, result.PerformerID, performer1.ID)
			assert.Equal(t, result.ApplicationID, application1.ID)
			assert.Equal(t, result.CreatedBy, customer1.ID)
			assert.Equal(t, result.ContractAddress, "")
			assert.Equal(t, result.CustomerAddress, customer1.EthereumAddress)
			assert.Equal(t, result.PerformerAddress, performer1.EthereumAddress)
			assert.True(t, decimal.RequireFromString("42.35").Equal(result.Price))
		}
	})
}

func TestGetContract(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/qwerty", nil)
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

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if contract does not exist", func(t *testing.T) {
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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/qwerty", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns success for authorized customer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractCreated,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ContractDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, model.ContractCreated, e.Status)
				assert.Equal(t, "Do it!", e.Title)
				assert.Equal(t, "Descriptive message", e.Description)
				assert.Equal(t, e.CustomerID, customer.ID)
				assert.Equal(t, e.PerformerID, performer.ID)
				assert.Equal(t, e.CreatedBy, customer.ID)
				assert.Equal(t, e.ApplicationID, application.ID)
				assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
				assert.EqualValues(t, 35, e.Duration)
			}
		}
	})

	t.Run("returns success for authorized performer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractCreated,
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, contractsURL+"/"+contract.ID, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.ContractDTO)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, model.ContractCreated, e.Status)
				assert.Equal(t, "Do it!", e.Title)
				assert.Equal(t, "Descriptive message", e.Description)
				assert.Equal(t, e.CustomerID, customer.ID)
				assert.Equal(t, e.PerformerID, performer.ID)
				assert.Equal(t, e.CreatedBy, customer.ID)
				assert.Equal(t, e.ApplicationID, application.ID)
				assert.True(t, decimal.RequireFromString("42.35").Equal(e.Price))
				assert.EqualValues(t, 35, e.Duration)
			}
		}
	})
}

func TestSetContractAsAccepted(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/accept", bytes.NewReader([]byte(body)))
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

	t.Run("returns error if contract does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/accept", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error for customer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/accept", bytes.NewReader([]byte(body)))
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

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/accept", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns success", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractCreated,
			ContractAddress: validBlockchainAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/accept", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			c, err := queries.ContractGet(ctx, contract.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, model.ContractAccepted, c.Status)
			}
		}
	})
}

func TestSetContractAsDeployed(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/deploy", bytes.NewReader([]byte(body)))
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

	t.Run("returns error if contract does not exist", func(t *testing.T) {
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
			"contract_address": "` + validBlockchainAddress + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if body is not a valid JSON", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{z}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
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

	t.Run("returns error for performer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"contract_address": "` + validBlockchainAddress + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("returns error if contract_address is missing", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address is required", e["message"])
		}
	})

	t.Run("returns error if contract_address is an empty string", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"contract_address": " "
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address is required", e["message"])
		}
	})

	t.Run("returns error if contract_address has an invalid format", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{
			"contract_address": "test"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address has an invalid format", e["message"])
		}
	})

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{
			"contract_address": "` + validBlockchainAddress + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if contract has an invalid status", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractDeployed,
		})
		require.NoError(t, err)

		body := `{
			"contract_address": "` + validBlockchainAddress + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "inappropriate action", e["message"])
			assert.Equal(t, "inappropriate action: unable to move from deployed to deployed", e["tech_info"])
		}
	})

	t.Run("returns error if deployed contract in blockchain has invalid values", func(t *testing.T) {
		t.Skip("Not implemented")
	})

	t.Run("returns success", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractAccepted,
		})
		require.NoError(t, err)

		body := `{
			"contract_address": "` + validBlockchainAddress + `"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/deploy", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			c, err := queries.ContractGet(ctx, contract.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, model.ContractDeployed, c.Status)
			}
		}
	})
}

func TestSetContractAsSigned(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/sign", bytes.NewReader([]byte(body)))
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

	t.Run("returns error if contract does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/sign", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error for customer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/sign", bytes.NewReader([]byte(body)))
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

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/sign", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if contract address has an invalid format", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractSigned,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/sign", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address has an invalid format", e["message"])
			assert.Equal(t, nil, e["tech_info"])
		}
	})

	t.Run("returns error if contract address has an invalid format", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractSigned,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/sign", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address has an invalid format", e["message"])
			assert.Equal(t, nil, e["tech_info"])
		}
	})

	t.Run("returns error if contract has an invalid status", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractSigned,
			ContractAddress: validBlockchainAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/sign", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "inappropriate action", e["message"])
			assert.Equal(t, "inappropriate action: unable to move from signed to signed", e["tech_info"])
		}
	})

	t.Run("returns error if deployed contract in blockchain has invalid values", func(t *testing.T) {
		t.Skip("Not implemented")
	})

	t.Run("returns success", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractDeployed,
			ContractAddress: validBlockchainAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/sign", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			c, err := queries.ContractGet(ctx, contract.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, model.ContractSigned, c.Status)
			}
		}
	})
}

func TestSetContractAsFunded(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/fund", bytes.NewReader([]byte(body)))
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

	t.Run("returns error if contract does not exist", func(t *testing.T) {
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

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error for performer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if contract address has an invalid format", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractFunded,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address has an invalid format", e["message"])
			assert.Equal(t, nil, e["tech_info"])
		}
	})

	t.Run("returns error if contract has an invalid status", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractFunded,
			ContractAddress: validBlockchainAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "inappropriate action", e["message"])
			assert.Equal(t, "inappropriate action: unable to move from funded to funded", e["tech_info"])
		}
	})

	t.Run("returns error if contract is not funded", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractSigned,
			ContractAddress: notFundedContractAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "the contract does not have sufficient funds", e["message"])
		}
	})

	t.Run("returns error if deployed contract in blockchain has invalid values", func(t *testing.T) {
		t.Skip("Not implemented")
	})

	t.Run("returns success", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "0.12",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractSigned,
			ContractAddress: fundedContractAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/fund", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			c, err := queries.ContractGet(ctx, contract.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, model.ContractFunded, c.Status)
			}
		}
	})
}

func TestSetContractAsApproved(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/approve", bytes.NewReader([]byte(body)))
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

	t.Run("returns error if contract does not exist", func(t *testing.T) {
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

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/approve", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error for performer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/approve", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/approve", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if contract address has an invalid format", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractFunded,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/approve", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address has an invalid format", e["message"])
			assert.Equal(t, nil, e["tech_info"])
		}
	})

	t.Run("returns error if contract has an invalid status", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractApproved,
			ContractAddress: validBlockchainAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/approve", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "inappropriate action", e["message"])
			assert.Equal(t, "inappropriate action: unable to move from approved to approved", e["tech_info"])
		}
	})

	t.Run("returns error if deployed contract in blockchain has invalid values", func(t *testing.T) {
		t.Skip("Not implemented")
	})

	t.Run("returns success", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "0.12",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractFunded,
			ContractAddress: fundedContractAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/approve", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			c, err := queries.ContractGet(ctx, contract.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, model.ContractApproved, c.Status)
			}
		}
	})
}

func TestSetContractAsCompleted(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/complete", bytes.NewReader([]byte(body)))
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

	t.Run("returns error if contract does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/qwerty/complete", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error for customer", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/complete", bytes.NewReader([]byte(body)))
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

	t.Run("returns error for another person", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/complete", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer2.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status)
	})

	t.Run("returns error if contract address has an invalid format", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			Title:         "Do it!",
			Description:   "Descriptive message",
			Price:         "42.35",
			Duration:      sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:    customer.ID,
			PerformerID:   performer.ID,
			ApplicationID: application.ID,
			CreatedBy:     customer.ID,
			Status:        model.ContractApproved,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/complete", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "contract_address has an invalid format", e["message"])
			assert.Equal(t, nil, e["tech_info"])
		}
	})

	t.Run("returns error if contract has an invalid status", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "42.35",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractCompleted,
			ContractAddress: validBlockchainAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/complete", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "inappropriate action", e["message"])
			assert.Equal(t, "inappropriate action: unable to move from completed to completed", e["tech_info"])
		}
	})

	t.Run("returns error if deployed contract in blockchain has invalid values", func(t *testing.T) {
		t.Skip("Not implemented")
	})

	t.Run("returns success", func(t *testing.T) {
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

		performer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
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
			Title:       "Contracts testing",
			Description: "Contracts testing description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Do it!",
			JobID:       job.ID,
			Price:       "42.35",
			ApplicantID: performer.ID,
		})
		require.NoError(t, err)

		contract, err := pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
			ID:              pgdao.NewID(),
			Title:           "Do it!",
			Description:     "Descriptive message",
			Price:           "0.12",
			Duration:        sql.NullInt32{Int32: 35, Valid: true},
			CustomerID:      customer.ID,
			PerformerID:     performer.ID,
			ApplicationID:   application.ID,
			CreatedBy:       customer.ID,
			Status:          model.ContractApproved,
			ContractAddress: fundedContractAddress,
		})
		require.NoError(t, err)

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, contractsURL+"/"+contract.ID+"/complete", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			c, err := queries.ContractGet(ctx, contract.ID)
			if assert.NoError(t, err) {
				assert.Equal(t, model.ContractCompleted, c.Status)
			}
		}
	})
}

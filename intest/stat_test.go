package intest

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

func TestStats(t *testing.T) {
	queries := pgdao.New(db)
	require.NoError(t, pgdao.PurgeDB(ctx, db))

	customer, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Date(2021, 8, 15, 14, 30, 45, 100, time.UTC),
	})
	require.NoError(t, err)

	applicant1, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	applicant2, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Date(2022, 1, 1, 3, 1, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	applicant3, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	applicant4, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	applicant5, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	applicant6, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	applicant7, err := queries.TestsPersonCreate(ctx, pgdao.TestsPersonCreateParams{
		ID:        pgdao.NewID(),
		Login:     pgdao.NewID(),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	job1, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Open Job",
		Description: "Description",
		CreatedBy:   customer.ID,
	})
	require.NoError(t, err)

	job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Another Open Job",
		Description: "Description",
		CreatedBy:   customer.ID,
	})
	require.NoError(t, err)

	suspendedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Suspended Job",
		Description: "Description",
		CreatedBy:   customer.ID,
	})
	require.NoError(t, err)

	err = pgdao.New(db).JobSuspend(ctx, suspendedJob.ID)
	require.NoError(t, err)

	blockedJob, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       "Blocked Job",
		Description: "Description",
		CreatedBy:   customer.ID,
	})
	require.NoError(t, err)

	err = pgdao.New(db).JobBlock(ctx, blockedJob.ID)
	require.NoError(t, err)

	application1, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 1",
		Price:       "42.35",
		JobID:       job1.ID,
		ApplicantID: applicant1.ID,
	})
	require.NoError(t, err)

	application2, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 2",
		Price:       "0.21",
		JobID:       job1.ID,
		ApplicantID: applicant2.ID,
	})
	require.NoError(t, err)

	application3, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 3",
		Price:       "0.22",
		JobID:       job1.ID,
		ApplicantID: applicant3.ID,
	})
	require.NoError(t, err)

	application4, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 4",
		Price:       "7.01",
		JobID:       job2.ID,
		ApplicantID: applicant4.ID,
	})
	require.NoError(t, err)

	application5, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 5",
		Price:       "8.33",
		JobID:       job2.ID,
		ApplicantID: applicant5.ID,
	})
	require.NoError(t, err)

	application6, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 6",
		Price:       "18.0",
		JobID:       job2.ID,
		ApplicantID: applicant6.ID,
	})
	require.NoError(t, err)

	application7, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     "Comment 7",
		Price:       "18.0",
		JobID:       job2.ID,
		ApplicantID: applicant7.ID,
	})
	require.NoError(t, err)

	// Created contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "10.1",
		CustomerID:    customer.ID,
		PerformerID:   applicant1.ID,
		ApplicationID: application1.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractCreated,
	})
	require.NoError(t, err)

	// Approved contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "22.003",
		CustomerID:    customer.ID,
		PerformerID:   applicant2.ID,
		ApplicationID: application2.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractApproved,
	})
	require.NoError(t, err)

	// Completed contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "99.999",
		CustomerID:    customer.ID,
		PerformerID:   applicant3.ID,
		ApplicationID: application3.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractCompleted,
	})
	require.NoError(t, err)

	// Accepted contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "99.999",
		CustomerID:    customer.ID,
		PerformerID:   applicant4.ID,
		ApplicationID: application4.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractAccepted,
	})
	require.NoError(t, err)

	// Deployed contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "99.999",
		CustomerID:    customer.ID,
		PerformerID:   applicant5.ID,
		ApplicationID: application5.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractDeployed,
	})
	require.NoError(t, err)

	// Signed contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "99.999",
		CustomerID:    customer.ID,
		PerformerID:   applicant6.ID,
		ApplicationID: application6.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractSigned,
	})
	require.NoError(t, err)

	// Funded contract
	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		Title:         "Our Contract",
		Description:   "Content",
		Price:         "99.999",
		CustomerID:    customer.ID,
		PerformerID:   applicant7.ID,
		ApplicationID: application7.ID,
		CreatedBy:     customer.ID,
		Status:        model.ContractFunded,
	})
	require.NoError(t, err)

	t.Run("returns stats", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL+"/stats", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			stats := new(model.Stats)

			require.NoError(t, json.Unmarshal(bb, stats))

			assert.Len(t, stats.RegistrationsByDate, 3)
			assert.EqualValues(t, 1, stats.RegistrationsByDate["2021-08-15"])
			assert.EqualValues(t, 1, stats.RegistrationsByDate["2022-01-01"])
			assert.EqualValues(t, 6, stats.RegistrationsByDate[time.Now().Format("2006-01-02")])

			assert.EqualValues(t, 8, stats.TotalRegistrations)
			assert.EqualValues(t, 2, stats.OpenedJobs)
			assert.EqualValues(t, 7, stats.TotalContracts)
			assert.True(t, decimal.RequireFromString("122.002").Equal(stats.TotalTransactionsVolume))
		}
	})
}

func TestStatsWithoutData(t *testing.T) {
	require.NoError(t, pgdao.PurgeDB(ctx, db))

	t.Run("returns empty stats", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL+"/stats", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			bb, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			stats := new(model.Stats)

			require.NoError(t, json.Unmarshal(bb, stats))

			assert.Len(t, stats.RegistrationsByDate, 0)
			assert.EqualValues(t, 0, stats.TotalRegistrations)
			assert.EqualValues(t, 0, stats.OpenedJobs)
			assert.EqualValues(t, 0, stats.TotalContracts)
			assert.True(t, decimal.Zero.Equal(stats.TotalTransactionsVolume))
		}
	})
}

package intest

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/service/pgsvc"

	_ "github.com/lib/pq"
)

var (
	dbURL  = os.Getenv("DB_URL")  // nolint: deadcode,varcheck
	appURL = os.Getenv("APP_URL") // nolint: deadcode,varcheck
)

var (
	db      *sql.DB
	queries *pgdao.Queries
	ctx     = context.Background()
)

func TestMain(m *testing.M) {
	makeExec, err := exec.LookPath("make")
	if err != nil {
		log.Fatal().Err(err).Msg("make not found")
	}

	cmd := exec.Command(makeExec, "migrate-drop", "migrate-up")

	cmd.Dir = "../pkg/db/db-migrations"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal().Err(err).Str("dbURL", dbURL).Msg("unable to open DB")
	}
	defer db.Close()

	queries = pgdao.New(db)

	if err := cmd.Run(); err != nil {
		log.Fatal().Err(err).Msg("unable to execute")
	}

	exitVal := m.Run()

	os.Exit(exitVal)
}

// Initialization functions

func addPerson(t *testing.T, login string) pgdao.Person {
	person, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        login,
		PasswordHash: pgsvc.CreateHashFromPassword(login + "-password"),
		DisplayName:  login,
		Email:        login + "@sample.com",
		AccessToken: sql.NullString{
			String: login + "-token",
			Valid:  true,
		},
	})
	require.NoError(t, err)

	return person
}

func addPersonWithEthereumAddress(t *testing.T, login, ethereum_address string) pgdao.Person {
	person, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        login,
		PasswordHash: pgsvc.CreateHashFromPassword(login + "-password"),
		DisplayName:  login,
		Email:        login + "@sample.com",
		AccessToken: sql.NullString{
			String: login + "-token",
			Valid:  true,
		},
		EthereumAddress: ethereum_address,
	})
	require.NoError(t, err)

	return person
}

func addJob(t *testing.T, title, description, createdBy, budget, duration string) pgdao.Job {
	durV, durE := strconv.Atoi(duration)

	job, err := queries.JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       title,
		Description: description,
		Budget: sql.NullString{
			String: budget,
			Valid:  budget != "",
		},
		Duration: sql.NullInt32{
			Int32: int32(durV),
			Valid: durE == nil,
		},
		CreatedBy: createdBy,
	})
	require.NoError(t, err)

	return job
}

func addApplication(t *testing.T, jobID, comment, price, applicantID string) pgdao.Application {
	application, err := queries.ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     comment,
		Price:       price,
		JobID:       jobID,
		ApplicantID: applicantID,
	})
	require.NoError(t, err)

	return application
}

func addContract(t *testing.T, customerID, performerID, applicationID, title, description, price, duration, customerAddress string) pgdao.Contract {
	durV, durE := strconv.Atoi(duration)

	contract, err := queries.ContractAdd(ctx, pgdao.ContractAddParams{
		ID:            pgdao.NewID(),
		CustomerID:    customerID,
		PerformerID:   performerID,
		ApplicationID: applicationID,
		Title:         title,
		Description:   description,
		Price:         price,
		Duration: sql.NullInt32{
			Int32: int32(durV),
			Valid: durE == nil,
		},
		CustomerAddress: customerAddress,
		CreatedBy:       customerID,
	})
	require.NoError(t, err)

	return contract
}

// HTTP stuff

func doRequest[T any](t *testing.T, httpMethod, url, body, token string) T {
	req, err := http.NewRequestWithContext(ctx, httpMethod, url, bytes.NewBufferString(body))
	require.NoError(t, err)

	req.Header.Set(clog.HeaderXHint, t.Name())
	req.Header.Set(echo.HeaderContentType, "application/json")
	if len(token) > 0 {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Truef(t, res.StatusCode >= 200 && res.StatusCode < 300, httpMethod+" "+url+": status should be 2xx. But actual is '%s'", res.Status)

	bb, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	resultObj := new(T)

	val := reflect.ValueOf(resultObj)

	if reflect.TypeOf([]byte{}).ConvertibleTo(val.Elem().Type()) {
		return reflect.ValueOf(bb).Convert(val.Elem().Type()).Interface().(T) // Don't ask questions, I'll explain everything later.
	} else {
		require.NoError(t, json.Unmarshal(bb, resultObj))
	}

	return *resultObj
}

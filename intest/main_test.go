package intest

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"testing"

	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq"
)

var (
	dbURL  = os.Getenv("DB_URL")  // nolint: deadcode,varcheck
	appURL = os.Getenv("APP_URL") // nolint: deadcode,varcheck
)

var (
	db  *sql.DB
	ctx = context.Background()
)

// token of default user
var testUserToken = ""

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

	if err := cmd.Run(); err != nil {
		log.Fatal().Err(err).Msg("unable to execute")
	}

	exitVal := m.Run()

	os.Exit(exitVal)
}

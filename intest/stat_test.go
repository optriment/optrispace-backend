package intest

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/pgsvc"
)

func TestStatRegistrations(t *testing.T) {
	queries := pgdao.New(db)
	require.NoError(t, pgdao.PurgeDB(ctx, db))

	_, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        pgdao.NewID(),
		PasswordHash: pgsvc.CreateHashFromPassword("1234"),
		DisplayName:  faker.New().Person().Name(),
		Email:        faker.New().Person().Contact().Email,
	})
	require.NoError(t, err)

	_, err = queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        pgdao.NewID(),
		PasswordHash: pgsvc.CreateHashFromPassword("1234"),
		DisplayName:  faker.New().Person().Name(),
		Email:        faker.New().Person().Contact().Email,
	})
	require.NoError(t, err)

	_, err = queries.PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        "inhouse",
		Login:        pgdao.NewID(),
		PasswordHash: pgsvc.CreateHashFromPassword("1234"),
		DisplayName:  faker.New().Person().Name(),
		Email:        faker.New().Person().Contact().Email,
	})
	require.NoError(t, err)

	type registrations struct {
		Day           string `json:"day"`
		Registrations int    `json:"registrations"`
	}

	t.Run("for 3 persons", func(t *testing.T) {
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

			assert.Len(t, stats.Registrations, 1)
			assert.EqualValues(t, 3, stats.Registrations[time.Now().Format("2006-01-02")])
		}
	})
}

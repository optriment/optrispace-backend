package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/db/pgdao"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatal().Err(err).Send()
	}
}

var (
	srcDBURL = os.Getenv("SRC_DB_URL")
	dstDBURL = os.Getenv("DST_DB_URL")
	force    = os.Getenv("FORCE") == "true"
)

func doMain() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srcDB, err := sql.Open("postgres", srcDBURL)
	if err != nil {
		return fmt.Errorf("unable to open source DB: %w", err)
	}
	defer srcDB.Close()

	dstDB, err := sql.Open("postgres", dstDBURL)
	if err != nil {
		return fmt.Errorf("unable to open destination DB: %w", err)
	}
	defer dstDB.Close()

	srcQueries := pgdao.New(srcDB)
	dstQueries := pgdao.New(dstDB)

	srcPersons, err := srcQueries.PersonsList(ctx)
	if err != nil {
		return fmt.Errorf("src: failed to list persons: %w", err)
	}
	log.Info().Int("n", len(srcPersons)).Msg("Read persons from src DB")

	for _, sp := range srcPersons {
		_, err := dstQueries.PersonGetByLogin(ctx, pgdao.PersonGetByLoginParams{
			Login: sp.Login,
			Realm: sp.Realm,
		})

		if err == nil {
			log.Info().Msgf("User '%s' already exists, ignored", sp.Login)
			continue
		}

		if errors.Is(err, sql.ErrNoRows) {
			log.Info().Bool("force", force).Msgf("Saving '%s' to DB", sp.Login)
			if force {
				dstQueries.PersonAddFull(ctx, pgdao.PersonAddFullParams{
					ID:              sp.ID,
					Realm:           sp.Realm,
					Login:           sp.Login,
					PasswordHash:    sp.PasswordHash,
					DisplayName:     sp.DisplayName,
					CreatedAt:       sp.CreatedAt,
					Email:           sp.Email,
					EthereumAddress: sp.EthereumAddress,
					Resources:       sp.Resources,
					AccessToken:     sp.AccessToken,
					IsAdmin:         sp.IsAdmin,
				})
			}
		} else {
			log.Warn().Err(err).Msgf("Failed to get info about user '%s', ignored", sp.Login)
		}

	}

	return nil
}

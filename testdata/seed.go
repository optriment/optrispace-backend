package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/jaswdr/faker"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/service/pgsvc"
)

var dbURL = os.Getenv("DB_URL")

func main() {
	ctx := context.Background()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal().Err(err).Str("dbURL", dbURL).Msg("unable to open DB")
	}

	defer db.Close()

	err = pgdao.PurgeDB(ctx, db)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to purge DB")
	}

	realm := "inhouse"
	defaultPassword := "password"
	defaultPasswordHash := pgsvc.CreateHashFromPassword(defaultPassword)

	faker := faker.New()

	user, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        realm,
		Login:        "admin",
		PasswordHash: defaultPasswordHash,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create user")
	}

	fmt.Printf("Created user with login: %s and password: %s\n", user.Login, defaultPassword)

	queries := pgdao.New(db)
	err = queries.PersonSetIsAdmin(ctx, pgdao.PersonSetIsAdminParams{
		IsAdmin: true,
		ID:      user.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to set user as admin")
	}

	fmt.Printf("Set user %s as admin\n", user.Login)

	customer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        realm,
		Login:        "customer1",
		PasswordHash: defaultPasswordHash,
		DisplayName:  faker.Person().Name(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create customer1")
	}

	fmt.Printf("Created customer with login: %s and password: %s\n", customer1.Login, defaultPassword)

	customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        realm,
		Login:        "customer2",
		PasswordHash: defaultPasswordHash,
		DisplayName:  faker.Person().Name(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create customer2")
	}

	fmt.Printf("Created customer with login: %s and password: %s\n", customer2.Login, defaultPassword)

	freelancer1, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        realm,
		Login:        "freelancer1",
		PasswordHash: defaultPasswordHash,
		DisplayName:  faker.Person().Name(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create freelancer1")
	}

	fmt.Printf("Created freelancer with login: %s and password: %s\n", freelancer1.Login, defaultPassword)

	freelancer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        realm,
		Login:        "freelancer2",
		PasswordHash: defaultPasswordHash,
		DisplayName:  faker.Person().Name(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create freelancer2")
	}

	fmt.Printf("Created freelancer with login: %s and password: %s\n", freelancer2.Login, defaultPassword)

	freelancer3, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
		ID:           pgdao.NewID(),
		Realm:        realm,
		Login:        "freelancer3",
		PasswordHash: defaultPasswordHash,
		DisplayName:  faker.Person().Name(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create freelancer3")
	}

	fmt.Printf("Created freelancer with login: %s and password: %s\n", freelancer3.Login, defaultPassword)

	job1, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       faker.Lorem().Sentence(faker.RandomNumber(1)),
		Description: faker.Lorem().Sentence(100),
		Budget: sql.NullString{
			String: fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
			Valid:  true,
		},
		Duration: sql.NullInt32{
			Int32: faker.Int32Between(0, 365),
			Valid: true,
		},
		CreatedBy: customer1.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create job1")
	}

	job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
		ID:          pgdao.NewID(),
		Title:       faker.Lorem().Sentence(faker.RandomNumber(1)),
		Description: faker.Lorem().Sentence(100),
		Budget: sql.NullString{
			String: fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
			Valid:  true,
		},
		Duration: sql.NullInt32{
			Int32: faker.Int32Between(0, 365),
			Valid: true,
		},
		CreatedBy: customer2.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create job2")
	}

	application1, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     faker.Lorem().Sentence(faker.RandomNumber(1)),
		Price:       fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		JobID:       job1.ID,
		ApplicantID: freelancer1.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create application1")
	}

	_, err = pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     faker.Lorem().Sentence(32),
		Price:       fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		JobID:       job1.ID,
		ApplicantID: freelancer2.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create application2")
	}

	_, err = pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     faker.Lorem().Sentence(32),
		Price:       fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		JobID:       job2.ID,
		ApplicantID: freelancer1.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create application3")
	}

	_, err = pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     faker.Lorem().Sentence(32),
		Price:       fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		JobID:       job2.ID,
		ApplicantID: freelancer2.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create application4")
	}

	application5, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
		ID:          pgdao.NewID(),
		Comment:     faker.Lorem().Sentence(32),
		Price:       fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		JobID:       job2.ID,
		ApplicantID: freelancer3.ID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create application5")
	}

	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:              pgdao.NewID(),
		Title:           faker.Lorem().Sentence(32),
		Description:     faker.Lorem().Sentence(100),
		Price:           fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		Duration:        sql.NullInt32{Int32: 42, Valid: true},
		CustomerID:      customer1.ID,
		PerformerID:     freelancer1.ID,
		ApplicationID:   application1.ID,
		CreatedBy:       customer1.ID,
		CustomerAddress: "0xDEADBEEF",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create contract1")
	}

	_, err = pgdao.New(db).ContractAdd(ctx, pgdao.ContractAddParams{
		ID:              pgdao.NewID(),
		Title:           faker.Lorem().Sentence(32),
		Description:     faker.Lorem().Sentence(100),
		Price:           fmt.Sprintf("%f", faker.RandomFloat(18, 1, 100)),
		Duration:        sql.NullInt32{Int32: 42, Valid: true},
		CustomerID:      customer2.ID,
		PerformerID:     freelancer3.ID,
		ApplicationID:   application5.ID,
		CreatedBy:       customer2.ID,
		CustomerAddress: "0xDEADBEEF",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("unable to create contract2")
	}
}

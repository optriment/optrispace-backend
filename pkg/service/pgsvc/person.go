package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jaswdr/faker"
	"github.com/lib/pq"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// PersonSvc is a person service
	PersonSvc struct {
		db *sql.DB
	}
)

// NewPerson creates service
func NewPerson(db *sql.DB) *PersonSvc {
	return &PersonSvc{db: db}
}

// Add implements service.Person
func (s *PersonSvc) Add(ctx context.Context, person *model.Person) (*model.Person, error) {
	var result *model.Person
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		input := pgdao.PersonAddParams{
			ID:           pgdao.NewID(),
			Realm:        person.Realm,
			Login:        person.Login,
			PasswordHash: CreateHashFromPassword(person.Password),
			DisplayName:  person.DisplayName,
			Email:        person.Email,
		}

		if input.Realm == "" {
			input.Realm = model.InhouseRealm
		}

		f := faker.New()

		if input.Login == "" {
			input.Login = input.ID
		}

		if input.DisplayName == "" {
			input.DisplayName = f.Person().Name()
		}

		o, err := queries.PersonAdd(ctx, input)

		if pqe, ok := err.(*pq.Error); ok { //nolint: errorlint
			if pqe.Code == "23505" {
				return fmt.Errorf("%s: %w", pqe.Detail, model.ErrDuplication)
			}
		}

		if err != nil {
			return fmt.Errorf("unable to PersonAdd: %w", err)
		}

		result = &model.Person{
			ID:          o.ID,
			Realm:       o.Realm,
			Login:       o.Login,
			DisplayName: o.DisplayName,
			CreatedAt:   o.CreatedAt,
			Email:       o.Email,
		}

		return nil
	})
}

// Get implements service.Person
func (s *PersonSvc) Get(ctx context.Context, id string) (*model.Person, error) {
	var result *model.Person
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		o, err := queries.PersonGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to PersonGet: %w", err)
		}

		result = &model.Person{
			ID:          o.ID,
			Realm:       o.Realm,
			Login:       o.Login,
			DisplayName: o.DisplayName,
			CreatedAt:   o.CreatedAt,
			Email:       o.Email,
		}

		return nil
	})
}

// List implements service.Person
func (s *PersonSvc) List(ctx context.Context) ([]*model.Person, error) {
	result := make([]*model.Person, 0)
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		oo, err := queries.PersonsList(ctx)
		if err != nil {
			return err
		}

		for _, o := range oo {
			result = append(result, &model.Person{
				ID:          o.ID,
				Realm:       o.Realm,
				Login:       o.Login,
				DisplayName: o.DisplayName,
				CreatedAt:   o.CreatedAt,
				Email:       o.Email,
			})
		}

		return nil
	})
}

// UpdatePassword implements service.Person
func (s *PersonSvc) UpdatePassword(ctx context.Context, subjectID, oldPassword, newPassword string) error {
	return doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		o, err := queries.PersonGet(ctx, subjectID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if e := CompareHashAndPassword(o.PasswordHash, oldPassword); e != nil {
			return model.ErrUnauthorized // for security reasons, we do not provide an exact reason of failure
		}

		return queries.PersonChangePassword(ctx, pgdao.PersonChangePasswordParams{
			NewPasswordHash: CreateHashFromPassword(newPassword),
			ID:              subjectID,
		})
	})
}

// Patch implements service.Person
func (s *PersonSvc) Patch(ctx context.Context, id, actorID string, patch map[string]any) error {
	return doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		if id != actorID {
			return model.ErrInsufficientRights
		}

		params := pgdao.PersonPatchParams{
			EthereumAddressChange: false,
			EthereumAddress:       "",
			ID:                    id,
		}

		v, c := patch["ethereum_address"]
		params.EthereumAddress, params.EthereumAddressChange = fmt.Sprint(v), c

		_, err := queries.PersonPatch(ctx, params)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		return err
	})
}

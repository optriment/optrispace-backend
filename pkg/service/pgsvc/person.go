package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

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
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		}

		if input.Realm == "" {
			input.Realm = model.InhouseRealm
		}

		input.Email = strings.ToLower(strings.TrimSpace(input.Email))

		input.DisplayName = strings.TrimSpace(input.DisplayName)

		if input.DisplayName == "" {
			now := time.Now()
			input.DisplayName = fmt.Sprintf("Person%d", now.Unix())
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

		result = personDBtoModel(o)

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

		result = personDBtoModel(o)

		return nil
	})
}

func personDBtoModel(o pgdao.Person) *model.Person {
	return &model.Person{
		ID:              o.ID,
		Realm:           o.Realm,
		Login:           o.Login,
		DisplayName:     o.DisplayName,
		CreatedAt:       o.CreatedAt,
		Email:           o.Email,
		Resources:       string(o.Resources),
		AccessToken:     o.AccessToken.String,
		IsAdmin:         o.IsAdmin,
		EthereumAddress: o.EthereumAddress,
	}
}

// GetByAccessToken implements service.Person
func (s *PersonSvc) GetByAccessToken(ctx context.Context, accessToken string) (*model.Person, error) {
	var result *model.Person
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		o, err := queries.PersonGetByAccessToken(ctx, accessToken)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to PersonGet: %w", err)
		}

		result = personDBtoModel(o)

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

		return queries.PersonSetPassword(ctx, pgdao.PersonSetPasswordParams{
			NewPasswordHash: CreateHashFromPassword(newPassword),
			ID:              subjectID,
		})
	})
}

// Patch implements service.Person
func (s *PersonSvc) Patch(ctx context.Context, id, actorID string, patch map[string]any) (*model.BasicPersonDTO, error) {
	var result *model.BasicPersonDTO

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		if id != actorID {
			return model.ErrInsufficientRights
		}

		params := pgdao.PersonPatchParams{
			EthereumAddressChange: false,
			EthereumAddress:       "",
			DisplayNameChange:     false,
			DisplayName:           "",
			EmailChange:           false,
			Email:                 "",
			ID:                    id,
		}

		v, c := patch["ethereum_address"]
		params.EthereumAddress, params.EthereumAddressChange = fmt.Sprint(v), c

		v, c = patch["display_name"]
		newDisplayName := strings.TrimSpace(fmt.Sprint(v))
		if newDisplayName != "" {
			params.DisplayName, params.DisplayNameChange = newDisplayName, c
		}

		v, c = patch["email"]
		newEmail := strings.ToLower(strings.TrimSpace(fmt.Sprint(v)))
		if newEmail != "" {
			params.Email, params.EmailChange = newEmail, c
		}

		o, err := queries.PersonPatch(ctx, params)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return err
		}

		result = &model.BasicPersonDTO{
			ID:              o.ID,
			Login:           o.Login,
			DisplayName:     o.DisplayName,
			Email:           o.Email,
			EthereumAddress: o.EthereumAddress,
			Resources:       string(o.Resources),
		}

		return nil
	})
}

// SetResources implements service.Person
func (s *PersonSvc) SetResources(ctx context.Context, id, actorID string, resources []byte) error {
	return doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		if id != actorID {
			return model.ErrInsufficientRights
		}

		return queries.PersonSetResources(ctx, pgdao.PersonSetResourcesParams{
			Resources: resources,
			ID:        id,
		})
	})
}

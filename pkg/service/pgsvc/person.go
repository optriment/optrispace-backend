package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
		person.ID = pgdao.NewID()

		p, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:      person.ID,
			Address: person.Address,
		})
		if err != nil {
			return fmt.Errorf("unable to PersonAdd: %w", err)
		}

		result = &model.Person{
			ID:        p.ID,
			Address:   p.Address,
			CreatedAt: p.CreatedAt,
		}

		return nil
	})
}

// Get implements service.Person
func (s *PersonSvc) Get(ctx context.Context, id string) (*model.Person, error) {
	var result *model.Person
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		p, err := queries.PersonGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to PersonGet: %w", err)
		}

		result = &model.Person{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			Address:   p.Address,
		}

		return nil
	})
}

// List implements service.Person
func (s *PersonSvc) List(ctx context.Context) ([]*model.Person, error) {
	result := make([]*model.Person, 0)
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		pp, err := queries.PersonsList(ctx)
		if err != nil {
			return err
		}

		for _, p := range pp {
			result = append(result, &model.Person{
				ID:        p.ID,
				CreatedAt: p.CreatedAt,
				Address:   p.Address,
			})
		}

		return nil
	})
}

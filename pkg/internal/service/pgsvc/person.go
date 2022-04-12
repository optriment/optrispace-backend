package pgsvc

import (
	"context"
	"database/sql"
	"fmt"

	"optrispace.com/work/pkg/internal/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	personSvc struct {
		db *sql.DB
	}
)

func NewPerson(db *sql.DB) *personSvc {
	return &personSvc{db: db}
}

// Add implements service.Person
func (s *personSvc) Add(ctx context.Context, person *model.Person) (*model.Person, error) {
	var result *model.Person
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		if person.ID == "" {
			person.ID = newID()
		}

		p, err := queries.PersonAdd(ctx, pgdao.PersonAddParams{
			ID:      person.ID,
			Address: person.Address,
		})
		if err != nil {
			return fmt.Errorf("unable to PersonAdd: %w", err)
		}

		result = &model.Person{
			ID:      p.ID,
			Address: p.Address,
		}

		return nil
	})
}

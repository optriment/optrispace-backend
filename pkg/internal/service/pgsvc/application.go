package pgsvc

import (
	"context"
	"database/sql"
	"fmt"

	"optrispace.com/work/pkg/internal/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	applicationSvc struct {
		db *sql.DB
	}
)

func NewApplication(db *sql.DB) *applicationSvc {
	return &applicationSvc{db: db}
}

func (s *applicationSvc) Add(ctx context.Context, jobID, applicantID string) (*model.Application, error) {
	var result *model.Application
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		id := newID()

		appl, err := queries.ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          id,
			JobID:       jobID,
			ApplicantID: applicantID,
		})
		if err != nil {
			return fmt.Errorf("unable to ApplicationAdd: %w", err)
		}

		result = &model.Application{
			ID:           appl.ID,
			CreationTime: appl.CreationTs,
			Applicant:    &model.Person{ID: appl.ApplicantID},
			Job:          &model.Job{ID: appl.JobID},
		}

		return nil
	})
}

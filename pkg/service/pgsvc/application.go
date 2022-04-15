package pgsvc

import (
	"context"
	"database/sql"
	"fmt"

	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// ApplicationSvc is a application service
	ApplicationSvc struct {
		db *sql.DB
	}
)

// NewApplication creates service
func NewApplication(db *sql.DB) *ApplicationSvc {
	return &ApplicationSvc{db: db}
}

// Add implements Application interface
func (s *ApplicationSvc) Add(ctx context.Context, jobID, applicantID string) (*model.Application, error) {
	var result *model.Application
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		id := pgdao.NewID()

		appl, err := queries.ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          id,
			JobID:       jobID,
			ApplicantID: applicantID,
		})
		if err != nil {
			return fmt.Errorf("unable to ApplicationAdd: %w", err)
		}

		result = &model.Application{
			ID:        appl.ID,
			CreatedAt: appl.CreationTs,
			Applicant: &model.Person{ID: appl.ApplicantID},
			Job:       &model.Job{ID: appl.JobID},
		}

		return nil
	})
}

package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
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
func (s *ApplicationSvc) Add(ctx context.Context, application *model.Application) (*model.Application, error) {
	var result *model.Application
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		id := pgdao.NewID()

		appl, err := queries.ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          id,
			Comment:     application.Comment,
			Price:       application.Price.String(),
			JobID:       application.Job.ID,
			ApplicantID: application.Applicant.ID,
		})

		if pqe, ok := err.(*pq.Error); ok { //nolint: errorlint
			if pqe.Code == "23505" {
				return fmt.Errorf("%s: %w", pqe.Detail, model.ErrDuplication)
			}
		}

		if err != nil {
			return fmt.Errorf("unable to ApplicationAdd: %w", err)
		}

		result = &model.Application{
			ID:        appl.ID,
			CreatedAt: appl.CreatedAt,
			UpdatedAt: appl.UpdatedAt,
			Applicant: &model.Person{ID: appl.ApplicantID},
			Comment:   appl.Comment,
			Price:     decimal.RequireFromString(appl.Price),
			Job:       &model.Job{ID: appl.JobID},
		}

		return nil
	})
}

// Get implements Application interface
func (s *ApplicationSvc) Get(ctx context.Context, id string) (*model.Application, error) {
	var result *model.Application
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		a, err := queries.ApplicationGet(ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to ApplicationGet with id=%s: %w", id, err)
		}

		result = &model.Application{
			ID:        a.ID,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			Applicant: &model.Person{ID: a.ApplicantID},
			Comment:   a.Comment,
			Price:     decimal.RequireFromString(a.Price),
			Job:       &model.Job{ID: a.JobID},
		}

		return nil
	})
}

// List implements Application interface
func (s *ApplicationSvc) List(ctx context.Context) ([]*model.Application, error) {
	return s.listBy(ctx, "", "")
}

// ListBy implements Application interface
func (s *ApplicationSvc) ListBy(ctx context.Context, jobID, actorID string) ([]*model.Application, error) {
	return s.listBy(ctx, jobID, actorID)
}

func (s *ApplicationSvc) listBy(ctx context.Context, jobID, actorID string) ([]*model.Application, error) {
	var result []*model.Application = make([]*model.Application, 0)
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		aa, err := queries.ApplicationsListBy(ctx, pgdao.ApplicationsListByParams{
			JobID:   jobID,
			ActorID: actorID,
		})
		if err != nil {
			return fmt.Errorf("unable to ApplicationsListBy: %w", err)
		}

		for _, a := range aa {
			var contract *model.Contract
			if a.ContractID.Valid {
				contract = &model.Contract{ID: a.ContractID.String}
			}

			result = append(result, &model.Application{
				ID:        a.ID,
				CreatedAt: a.CreatedAt,
				UpdatedAt: a.UpdatedAt,
				Applicant: &model.Person{ID: a.ApplicantID},
				Comment:   a.Comment,
				Price:     decimal.RequireFromString(a.Price),
				Job:       &model.Job{ID: a.JobID},
				Contract:  contract,
			})
		}

		return nil
	})
}

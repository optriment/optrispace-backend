package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ryboe/q"
	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// JobSvc is a job service
	JobSvc struct {
		db *sql.DB
	}
)

// NewJob creates service
func NewJob(db *sql.DB) *JobSvc {
	return &JobSvc{db: db}
}

// Add implements service.Job interface
func (s *JobSvc) Add(ctx context.Context, e *model.Job) (*model.Job, error) {
	var result *model.Job
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		j, err := queries.JobAdd(ctx,
			pgdao.JobAddParams{
				ID:          pgdao.NewID(),
				Title:       e.Title,
				Description: e.Description,
				Budget: sql.NullString{
					String: e.Budget.String(),
					Valid:  !e.Budget.Equal(decimal.Zero),
				},
				Duration: sql.NullInt32{
					Int32: e.Duration,
					Valid: e.Duration > 0,
				},
				CreatedBy: e.CreatedBy.ID,
			})
		if err != nil {
			return fmt.Errorf("unable to JobAdd job: %w", err)
		}

		budget := decimal.Zero
		if j.Budget.Valid {
			budget = decimal.RequireFromString(j.Budget.String)
		}

		result = &model.Job{
			ID:          j.ID,
			Title:       j.Title,
			Description: j.Description,
			Budget:      budget,
			Duration:    j.Duration.Int32,
			CreatedAt:   j.CreatedAt,
			CreatedBy: &model.Person{
				ID: j.CreatedBy,
			},
			UpdatedAt: j.UpdatedAt,
		}
		return nil
	})
}

// Get implements service.Job interface
func (s *JobSvc) Get(ctx context.Context, id string) (*model.Job, error) {
	var result *model.Job
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		j, err := queries.JobGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet jobs with id='%s': %w", id, err)
		}

		budget := decimal.Zero
		if j.Budget.Valid {
			budget = decimal.RequireFromString(j.Budget.String)
		}

		result = &model.Job{
			ID:          j.ID,
			Title:       j.Title,
			Description: j.Description,
			Budget:      budget,
			Duration:    j.Duration.Int32,
			CreatedAt:   j.CreatedAt,
			CreatedBy:   &model.Person{ID: j.CreatedBy},
			UpdatedAt:   j.UpdatedAt,
		}
		return nil
	})
}

// List implements service.Job interface
func (s *JobSvc) List(ctx context.Context) ([]*model.Job, error) {
	var result []*model.Job = make([]*model.Job, 0)
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		jj, err := queries.JobsList(ctx)
		q.Q(jj, err)
		if err != nil {
			return fmt.Errorf("unable to JobReadAll job: %w", err)
		}
		for _, j := range jj {
			budget := decimal.Zero
			if j.Budget.Valid {
				budget = decimal.RequireFromString(j.Budget.String)
			}

			result = append(result, &model.Job{
				ID:          j.ID,
				Title:       j.Title,
				Description: j.Description,
				Budget:      budget,
				Duration:    j.Duration.Int32,
				CreatedAt:   j.CreatedAt,
				CreatedBy:   &model.Person{ID: j.CreatedBy},
				UpdatedAt:   j.UpdatedAt,
			})
		}
		return nil
	})
}

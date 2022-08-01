package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
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
		if e.Budget.IsNegative() {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: model.ValidationErrorMustBePositive("budget"),
			}
		}

		o, err := queries.JobAdd(ctx,
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
				CreatedBy: e.CreatedBy,
			})
		if err != nil {
			return fmt.Errorf("unable to JobAdd job: %w", err)
		}

		budget := decimal.Zero
		if o.Budget.Valid {
			budget = decimal.RequireFromString(o.Budget.String)
		}

		result = &model.Job{
			ID:          o.ID,
			Title:       o.Title,
			Description: o.Description,
			Budget:      budget,
			Duration:    o.Duration.Int32,
			CreatedAt:   o.CreatedAt,
			CreatedBy:   o.CreatedBy,
			UpdatedAt:   o.UpdatedAt,
		}
		return nil
	})
}

// Get implements service.Job interface
func (s *JobSvc) Get(ctx context.Context, id string) (*model.Job, error) {
	var result *model.Job
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		o, err := queries.JobGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet with id='%s': %w", id, err)
		}

		budget := decimal.Zero
		if o.Budget.Valid {
			budget = decimal.RequireFromString(o.Budget.String)
		}

		result = &model.Job{
			ID:                o.ID,
			Title:             o.Title,
			Description:       o.Description,
			Budget:            budget,
			Duration:          o.Duration.Int32,
			CreatedAt:         o.CreatedAt,
			CreatedBy:         o.CreatedBy,
			UpdatedAt:         o.UpdatedAt,
			ApplicationsCount: uint(o.ApplicationCount),
			Customer: &model.Person{
				ID:          o.CreatedBy,
				DisplayName: o.CustomerDisplayName,
			},
		}
		return nil
	})
}

// List implements service.Job interface
func (s *JobSvc) List(ctx context.Context) ([]*model.Job, error) {
	result := make([]*model.Job, 0)
	return result, doWithQueries(ctx, s.db, defaultRoTxOpts, func(queries *pgdao.Queries) error {
		oo, err := queries.JobsList(ctx)
		if err != nil {
			return fmt.Errorf("unable to JobReadAll job: %w", err)
		}
		for _, o := range oo {
			budget := decimal.Zero
			if o.Budget.Valid {
				budget = decimal.RequireFromString(o.Budget.String)
			}

			result = append(result, &model.Job{
				ID:                o.ID,
				Title:             o.Title,
				Description:       o.Description,
				Budget:            budget,
				Duration:          o.Duration.Int32,
				CreatedAt:         o.CreatedAt,
				CreatedBy:         o.CreatedBy,
				UpdatedAt:         o.UpdatedAt,
				ApplicationsCount: uint(o.ApplicationCount),
			})
		}
		return nil
	})
}

// Patch implements service.Job interface
func (s *JobSvc) Patch(ctx context.Context, id, actorID string, patch map[string]any) (*model.Job, error) {
	var result *model.Job
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		params := &pgdao.JobPatchParams{
			ID:    id,
			Actor: actorID,
		}
		_, params.TitleChange = patch["title"]
		_, params.DescriptionChange = patch["description"]
		_, params.BudgetChange = patch["budget"]
		_, params.DurationChange = patch["duration"]

		err := mapstructure.Decode(patch, params)
		if err != nil {
			return fmt.Errorf("unable to decode patch from struct: %w", err)
		}

		if params.BudgetChange {
			if d, e := decimal.NewFromString(params.Budget); e != nil {
				return &model.BackendError{
					Cause:   model.ErrValidationFailed,
					Message: model.ValidationErrorInvalidFormat("budget"),
				}
			} else if d.IsNegative() {
				return &model.BackendError{
					Cause:   model.ErrValidationFailed,
					Message: model.ValidationErrorMustBePositive("budget"),
				}
			}
		}

		o, err := queries.JobPatch(ctx, *params)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobPatch with id='%s': %w", id, err)
		}

		budget := decimal.Zero
		if o.Budget.Valid {
			budget = decimal.RequireFromString(o.Budget.String)
		}

		result = &model.Job{
			ID:          o.ID,
			Title:       o.Title,
			Description: o.Description,
			Budget:      budget,
			Duration:    o.Duration.Int32,
			CreatedAt:   o.CreatedAt,
			CreatedBy:   o.CreatedBy,
			UpdatedAt:   o.UpdatedAt,
			Customer: &model.Person{
				ID: o.CreatedBy,
			},
		}
		return nil
	})
}

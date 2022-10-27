package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
func (s *JobSvc) Add(ctx context.Context, customerID string, dto *model.CreateJobDTO) (*model.JobDTO, error) {
	var result *model.JobDTO

	if strings.TrimSpace(dto.Title) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("title"),
		}
	}

	if strings.TrimSpace(dto.Description) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("description"),
		}
	}

	if dto.Budget.IsNegative() {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustBePositive("budget"),
		}
	}

	if dto.Duration < 0 {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustNotBeNegative("duration"),
		}
	}

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		customer, err := queries.PersonGet(ctx, customerID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "customer does not exist",
			}
		}

		customerEthereumAddress := strings.ToLower(strings.TrimSpace(customer.EthereumAddress))
		if customerEthereumAddress == "" {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "customer does not have wallet",
			}
		}

		jobParams := pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       strings.TrimSpace(dto.Title),
			Description: strings.TrimSpace(dto.Description),
			Budget: sql.NullString{
				String: dto.Budget.String(),
				Valid:  !dto.Budget.Equal(decimal.Zero),
			},
			Duration: sql.NullInt32{
				Int32: dto.Duration,
				Valid: dto.Duration > 0,
			},
			CreatedBy: customer.ID,
		}

		newJob, err := queries.JobAdd(ctx, jobParams)
		if err != nil {
			return fmt.Errorf("unable to JobAdd job: %w", err)
		}

		budget := decimal.Zero
		if newJob.Budget.Valid {
			budget = decimal.RequireFromString(newJob.Budget.String)
		}

		result = &model.JobDTO{
			ID:          newJob.ID,
			Title:       newJob.Title,
			Description: newJob.Description,
			Budget:      budget,
			Duration:    newJob.Duration.Int32,
			CreatedAt:   newJob.CreatedAt,
			UpdatedAt:   newJob.UpdatedAt,
			CreatedBy:   newJob.CreatedBy,
		}
		return nil
	})
}

// Get implements service.Job interface
func (s *JobSvc) Get(ctx context.Context, id string) (*model.JobCardDTO, error) {
	var result *model.JobCardDTO
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

		result = &model.JobCardDTO{}
		result.ID = o.ID
		result.Title = o.Title
		result.Description = o.Description
		result.Budget = budget
		result.Duration = o.Duration.Int32
		result.CreatedAt = o.CreatedAt
		result.CreatedBy = o.CreatedBy
		result.UpdatedAt = o.UpdatedAt
		result.ApplicationsCount = uint(o.ApplicationCount)
		result.CustomerDisplayName = o.CustomerDisplayName
		result.CustomerEthereumAddress = o.CustomerEthereumAddress
		result.IsSuspended = o.SuspendedAt.Valid

		return nil
	})
}

// List implements service.Job interface
func (s *JobSvc) List(ctx context.Context) ([]*model.JobDTO, error) {
	result := make([]*model.JobDTO, 0)
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

			result = append(result, &model.JobDTO{
				ID:                      o.ID,
				Title:                   o.Title,
				Description:             o.Description,
				Budget:                  budget,
				CreatedAt:               o.CreatedAt,
				CreatedBy:               o.CreatedBy,
				ApplicationsCount:       uint(o.ApplicationCount),
				CustomerDisplayName:     o.CustomerDisplayName,
				CustomerEthereumAddress: o.CustomerEthereumAddress,
			})
		}
		return nil
	})
}

// Block implements service.Job interface
func (s *JobSvc) Block(ctx context.Context, id, actorID string) error {
	return doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "person does not exist",
			}
		}

		if !person.IsAdmin {
			return model.ErrInsufficientRights
		}

		job, err := queries.JobGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet with id='%s': %w", id, err)
		}

		return queries.JobBlock(ctx, job.ID)
	})
}

// Suspend implements service.Job interface
func (s *JobSvc) Suspend(ctx context.Context, id, actorID string) error {
	return doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "person does not exist",
			}
		}

		job, err := queries.JobGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet with id='%s': %w", id, err)
		}

		if person.ID != job.CreatedBy {
			return model.ErrInsufficientRights
		}

		return queries.JobSuspend(ctx, job.ID)
	})
}

// Resume implements service.Job interface
func (s *JobSvc) Resume(ctx context.Context, id, actorID string) error {
	return doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "person does not exist",
			}
		}

		job, err := queries.JobGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet with id='%s': %w", id, err)
		}

		if person.ID != job.CreatedBy {
			return model.ErrInsufficientRights
		}

		if !job.SuspendedAt.Valid {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "job is not suspended",
			}
		}

		return queries.JobResume(ctx, job.ID)
	})
}

// Patch implements service.Job interface
func (s *JobSvc) Patch(ctx context.Context, id, actorID string, dto *model.UpdateJobDTO) (*model.JobDTO, error) {
	var result *model.JobDTO

	if strings.TrimSpace(dto.Title) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("title"),
		}
	}

	if strings.TrimSpace(dto.Description) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("description"),
		}
	}

	if dto.Budget.IsNegative() {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustBePositive("budget"),
		}
	}

	if dto.Duration < 0 {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustNotBeNegative("duration"),
		}
	}

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		customer, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "customer does not exist",
			}
		}

		customerEthereumAddress := strings.ToLower(strings.TrimSpace(customer.EthereumAddress))
		if customerEthereumAddress == "" {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "customer does not have wallet",
			}
		}

		params := &pgdao.JobPatchParams{
			ID:    id,
			Actor: customer.ID,
		}

		params.Title = strings.TrimSpace(dto.Title)
		params.Description = strings.TrimSpace(dto.Description)
		params.Budget = dto.Budget.String()
		params.Duration = dto.Duration

		_, err = queries.JobPatch(ctx, *params)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobPatch with id='%s': %w", id, err)
		}

		job, err := queries.JobGet(ctx, id)
		if err != nil {
			return fmt.Errorf("unable to JobGet with id='%s': %w", id, err)
		}

		result = &model.JobDTO{
			ID:                      job.ID,
			Title:                   job.Title,
			Description:             job.Description,
			Budget:                  decimal.RequireFromString(job.Budget.String),
			Duration:                job.Duration.Int32,
			CreatedAt:               job.CreatedAt,
			CreatedBy:               job.CreatedBy,
			UpdatedAt:               job.UpdatedAt,
			ApplicationsCount:       uint(job.ApplicationCount),
			CustomerDisplayName:     job.CustomerDisplayName,
			CustomerEthereumAddress: job.CustomerEthereumAddress,
		}

		return nil
	})
}

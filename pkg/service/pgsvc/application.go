package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/clog"
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

		if application.Price.IsNegative() {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: model.ValidationErrorMustBePositive("Price"),
			}
		}

		job, err := queries.JobGet(ctx, application.Job.ID)
		if errors.Is(err, sql.ErrNoRows) {
			return &model.BackendError{
				Cause:    model.ErrEntityNotFound,
				Message:  "job not found",
				TechInfo: application.Job.ID,
			}
		}
		if err != nil {
			return fmt.Errorf("unable to get job %s info: %w", application.Job.ID, err)
		}

		appl, err := queries.ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          id,
			Comment:     application.Comment,
			Price:       application.Price.String(),
			JobID:       application.Job.ID,
			ApplicantID: application.Applicant.ID,
		})

		if pqe, ok := err.(*pq.Error); ok { //nolint: errorlint
			if pqe.Code == "23505" {
				return fmt.Errorf("%s: %w", pqe.Detail, model.ErrApplicationAlreadyExists)
			}

			if pqe.Code == "23503" && pqe.Constraint == "applications_job_id_fkey" {
				return &model.BackendError{
					Cause:    model.ErrEntityNotFound,
					Message:  "job not found",
					TechInfo: application.Job.ID,
				}
			}
		}

		if err != nil {
			return fmt.Errorf("unable to ApplicationAdd: %w", err)
		}

		if _, e := newChat(ctx, queries, chatTopicApplication(appl.ID), appl.Comment, appl.ApplicantID, job.CreatedBy); e != nil {
			clog.Ctx(ctx).Warn().Err(e).Str("applicationID", appl.ID).Msg("Failed to create chat for application")
		}

		result = &model.Application{
			ID:        appl.ID,
			CreatedAt: appl.CreatedAt,
			UpdatedAt: appl.UpdatedAt,
			Applicant: &model.JobApplicant{ID: appl.ApplicantID},
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
			Applicant: &model.JobApplicant{ID: a.ApplicantID, DisplayName: a.ApplicantDisplayName, EthereumAddress: a.ApplicantEthereumAddress},
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
	result := make([]*model.Application, 0)
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		aa, err := queries.ApplicationsListBy(ctx, pgdao.ApplicationsListByParams{
			JobID:   jobID,
			ActorID: actorID,
		})
		if err != nil {
			return fmt.Errorf("unable to ApplicationsListBy: %w", err)
		}

		for _, a := range aa {
			job := &model.Job{
				ID:          a.JobID,
				Title:       a.JobTitle,
				Description: a.JobDescription,
			}

			if a.JobBudget.Valid {
				job.Budget = decimal.RequireFromString(a.JobBudget.String)
			}

			var contract *model.Contract

			if a.ContractID.Valid {
				contract = &model.Contract{
					ID:     a.ContractID.String,
					Status: a.ContractStatus.String,
					Price:  decimal.RequireFromString(a.ContractPrice.String),
				}
			}

			result = append(result, &model.Application{
				ID:        a.ID,
				CreatedAt: a.CreatedAt,
				UpdatedAt: a.UpdatedAt,
				Applicant: &model.JobApplicant{ID: a.ApplicantID, DisplayName: a.ApplicantDisplayName, EthereumAddress: a.ApplicantEthereumAddress},
				Comment:   a.Comment,
				Price:     decimal.RequireFromString(a.Price),
				Job:       job,
				Contract:  contract,
			})
		}

		return nil
	})
}

// ListByApplicant returns all applications for specific applicant
func (s *ApplicationSvc) ListByApplicant(ctx context.Context, applicantID string) ([]*model.Application, error) {
	result := make([]*model.Application, 0)
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		aa, err := queries.ApplicationsGetByApplicant(ctx, applicantID)
		if err != nil {
			return fmt.Errorf("unable to ApplicationsGetByApplicant: %w", err)
		}

		for _, a := range aa {
			job := &model.Job{
				ID:          a.JobID,
				Title:       a.JobTitle,
				Description: a.JobDescription,
			}

			if a.JobBudget.Valid {
				job.Budget = decimal.RequireFromString(a.JobBudget.String)
			}

			var contract *model.Contract

			if a.ContractID.Valid {
				contract = &model.Contract{
					ID:     a.ContractID.String,
					Status: a.ContractStatus.String,
					Price:  decimal.RequireFromString(a.ContractPrice.String),
				}
			}

			result = append(result, &model.Application{
				ID:        a.ID,
				CreatedAt: a.CreatedAt,
				UpdatedAt: a.UpdatedAt,
				Applicant: &model.JobApplicant{ID: a.ApplicantID},
				Comment:   a.Comment,
				Price:     decimal.RequireFromString(a.Price),
				Job:       job,
				Contract:  contract,
			})
		}

		return nil
	})
}

// GetChat implements Application interface
func (s *ApplicationSvc) GetChat(ctx context.Context, id, actorID string) (*model.Chat, error) {
	var result *model.Chat
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		application, err := queries.ApplicationGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			err = model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to get application by id %s: %w", id, err)
		}

		job, err := queries.JobGet(ctx, application.JobID)
		if errors.Is(err, sql.ErrNoRows) {
			err = model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to get job by id %s: %w", application.JobID, err)
		}

		if actorID != application.ApplicantID && actorID != job.CreatedBy {
			return fmt.Errorf("user %s has no rights to acquire chat info: %w", actorID, model.ErrEntityNotFound)
		}

		topic := chatTopicApplication(id)
		ec, err := queries.ChatGetByTopic(ctx, topic)

		if err == nil {
			result = chatFromDB(ec)
			return nil
		}

		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		result, err = newChat(ctx, queries, topic, application.Comment, application.ApplicantID, job.CreatedBy)

		return err
	})
}

package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
func (s *ApplicationSvc) Add(ctx context.Context, applicantID string, dto *model.CreateApplicationDTO) (*model.ApplicationDTO, error) {
	var result *model.ApplicationDTO

	if strings.TrimSpace(dto.Comment) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("comment"),
		}
	}

	if decimal.Zero.Equal(dto.Price) {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("price"),
		}
	}

	if dto.Price.IsNegative() {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustBePositive("price"),
		}
	}

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		job, err := queries.JobGet(ctx, dto.JobID)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to get job %s info: %w", dto.JobID, err)
		}

		if job.SuspendedAt.Valid {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "job does not accept new applications",
			}
		}

		applicant, err := queries.PersonGet(ctx, applicantID)
		if err != nil {
			return model.ErrInsufficientRights
		}

		if applicant.ID == job.CreatedBy {
			return model.ErrInsufficientRights
		}

		applicantEthereumAddress := strings.ToLower(strings.TrimSpace(applicant.EthereumAddress))
		if applicantEthereumAddress == "" {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "applicant does not have wallet",
			}
		}

		applicationParams := pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     strings.TrimSpace(dto.Comment),
			Price:       dto.Price.String(),
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		}

		newApplication, err := queries.ApplicationAdd(ctx, applicationParams)

		if pqe, ok := err.(*pq.Error); ok { //nolint: errorlint
			if pqe.Code == "23505" {
				return fmt.Errorf("%s: %w", pqe.Detail, model.ErrApplicationAlreadyExists)
			}
		}

		if err != nil {
			return fmt.Errorf("unable to ApplicationAdd: %w", err)
		}

		if _, e := newChat(ctx, queries, newChatTopicApplication(newApplication.ID), newApplication.Comment, newApplication.ApplicantID, job.CreatedBy); e != nil {
			clog.Ctx(ctx).Warn().Err(e).Str("applicationID", newApplication.ID).Msg("Failed to create chat for application")
		}

		result = &model.ApplicationDTO{
			ID:          newApplication.ID,
			JobID:       job.ID,
			ApplicantID: applicant.ID,
			Comment:     newApplication.Comment,
			Price:       decimal.RequireFromString(newApplication.Price),
			CreatedAt:   newApplication.CreatedAt,
		}

		return nil
	})
}

// GetForJob returns application for specific job by applicant
func (s *ApplicationSvc) GetForJob(ctx context.Context, jobID, actorID string) (*model.ApplicationDTO, error) {
	var result *model.ApplicationDTO
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		job, err := queries.JobGet(ctx, jobID)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to get job %s info: %w", jobID, err)
		}

		applicant, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return model.ErrInsufficientRights
		}

		if applicant.ID == job.CreatedBy {
			return model.ErrInsufficientRights
		}

		a, err := queries.ApplicationFindByJobAndApplicant(ctx, pgdao.ApplicationFindByJobAndApplicantParams{
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})

		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("unable to ApplicationFindByJobAndApplicant with jobID=%s, applicant_id=%s: %w", job.ID, applicant.ID, err)
		}

		result = &model.ApplicationDTO{
			ID:                       a.ID,
			JobID:                    a.JobID,
			ApplicantID:              a.ApplicantID,
			Comment:                  a.Comment,
			Price:                    decimal.RequireFromString(a.Price),
			CreatedAt:                a.CreatedAt,
			ApplicantEthereumAddress: a.ApplicantEthereumAddress,
			ApplicantDisplayName:     a.ApplicantDisplayName,
			ContractID:               a.ContractID.String,
			ContractStatus:           a.ContractStatus.String,
		}

		return nil
	})
}

// Get implements Application interface
func (s *ApplicationSvc) Get(ctx context.Context, id, actorID string) (*model.ApplicationDTO, error) {
	var result *model.ApplicationDTO
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		application, err := queries.ApplicationGet(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to ApplicationGet with id=%s: %w", id, err)
		}

		job, err := queries.JobGet(ctx, application.JobID)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet with id=%s: %w", id, err)
		}

		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return model.ErrInsufficientRights
		}

		if person.ID != application.ApplicantID && person.ID != job.CreatedBy {
			return model.ErrInsufficientRights
		}

		result = &model.ApplicationDTO{
			ID:                       application.ID,
			JobID:                    application.JobID,
			JobTitle:                 application.JobTitle,
			JobDescription:           application.JobDescription,
			ApplicantID:              application.ApplicantID,
			Comment:                  application.Comment,
			Price:                    decimal.RequireFromString(application.Price),
			CreatedAt:                application.CreatedAt,
			ApplicantEthereumAddress: application.ApplicantEthereumAddress,
			ApplicantDisplayName:     application.ApplicantDisplayName,
			ContractID:               application.ContractID.String,
			ContractStatus:           application.ContractStatus.String,
		}

		if application.JobBudget.Valid {
			result.JobBudget = decimal.RequireFromString(application.JobBudget.String)
		}

		return nil
	})
}

// ListByJob returns applications belong to specific job by ID
func (s *ApplicationSvc) ListByJob(ctx context.Context, jobID, actorID string) ([]*model.ApplicationDTO, error) {
	result := make([]*model.ApplicationDTO, 0)

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		job, err := queries.JobGet(ctx, jobID)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to JobGet with id='%s': %w", jobID, err)
		}

		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return model.ErrInsufficientRights
		}

		if person.ID != job.CreatedBy {
			return model.ErrInsufficientRights
		}

		aa, err := queries.ApplicationsGetByJob(ctx, job.ID)
		if err != nil {
			return fmt.Errorf("unable to ApplicationsGetByJob: %w", err)
		}

		for _, a := range aa {
			app := &model.ApplicationDTO{
				ID:                       a.ID,
				JobID:                    a.JobID,
				ApplicantID:              a.ApplicantID,
				Comment:                  a.Comment,
				Price:                    decimal.RequireFromString(a.Price),
				CreatedAt:                a.CreatedAt,
				ApplicantEthereumAddress: a.ApplicantEthereumAddress,
				ApplicantDisplayName:     a.ApplicantDisplayName,
				ContractID:               a.ContractID.String,
				ContractStatus:           a.ContractStatus.String,
			}

			result = append(result, app)
		}

		return nil
	})
}

// ListByApplicant returns all applications for specific applicant
func (s *ApplicationSvc) ListByApplicant(ctx context.Context, actorID string) ([]*model.ApplicationDTO, error) {
	result := make([]*model.ApplicationDTO, 0)
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return model.ErrInsufficientRights
		}

		aa, err := queries.ApplicationsGetByApplicant(ctx, person.ID)
		if err != nil {
			return fmt.Errorf("unable to ApplicationsGetByApplicant: %w", err)
		}

		for _, a := range aa {
			app := &model.ApplicationDTO{
				ID:                       a.ID,
				JobID:                    a.JobID,
				JobTitle:                 a.JobTitle,
				JobDescription:           a.JobDescription,
				ApplicantID:              a.ApplicantID,
				Comment:                  a.Comment,
				Price:                    decimal.RequireFromString(a.Price),
				CreatedAt:                a.CreatedAt,
				ApplicantEthereumAddress: a.ApplicantEthereumAddress,
				ApplicantDisplayName:     a.ApplicantDisplayName,
				ContractID:               a.ContractID.String,
				ContractStatus:           a.ContractStatus.String,
			}

			if a.JobBudget.Valid {
				app.JobBudget = decimal.RequireFromString(a.JobBudget.String)
			}

			result = append(result, app)
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

		person, err := queries.PersonGet(ctx, actorID)
		if err != nil {
			return model.ErrInsufficientRights
		}

		job, err := queries.JobGet(ctx, application.JobID)

		if errors.Is(err, sql.ErrNoRows) {
			err = model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to get job by id %s: %w", application.JobID, err)
		}

		if person.ID != application.ApplicantID && person.ID != job.CreatedBy {
			return model.ErrInsufficientRights
		}

		topic := newChatTopicApplication(id)
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

package service

import (
	"context"
	"database/sql"

	"optrispace.com/work/pkg/internal/service/pgsvc"
	"optrispace.com/work/pkg/model"
)

type (
	// Job handles job offers
	Job interface {
		// Add saves the job offer into storage
		Add(ctx context.Context, job *model.Job) (*model.Job, error)
		// Get reads specified job from storage by specified id
		// It can return model.ErrNotFound
		Get(ctx context.Context, id string) (*model.Job, error)

		// List reads all items from storage and returns result to
		List(ctx context.Context) ([]*model.Job, error)
	}

	// Person is a person who pay or earn
	Person interface {
		// Add creates new person
		Add(ctx context.Context, person *model.Person) (*model.Person, error)
	}

	// Application is application for a job offer
	Application interface {
		// Add creates new connection for applicant and job offer
		Add(ctx context.Context, jobID, applicantID string) (*model.Application, error)
	}
)

func NewJob(db *sql.DB) Job {
	return pgsvc.NewJob(db)
}

func NewPerson(db *sql.DB) Person {
	return pgsvc.NewPerson(db)
}

func NewApplication(db *sql.DB) Application {
	return pgsvc.NewApplication(db)
}

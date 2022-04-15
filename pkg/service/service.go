package service

import (
	"context"
	"database/sql"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/pgsvc"
)

type (

	// Security creates user representation from echo.Context, if there is such data
	Security interface {
		// FromContext acquires user from echo and persisten storage
		// It will return never nil
		FromContext(c echo.Context) (*model.UserContext, error)
	}

	// GenericCRUD represents the standart methods for CRUD operations
	GenericCRUD[E any] interface {
		// Add saves the entity into storage
		Add(ctx context.Context, e *E) (*E, error)
		// Get reads specified entity from storage by specified id
		// It can return model.ErrNotFound
		Get(ctx context.Context, id string) (*E, error)

		// List reads all items from storage
		List(ctx context.Context) ([]*E, error)
	}

	// Job handles job offers
	Job interface {
		GenericCRUD[model.Job]
	}

	// Person is a person who pay or earn
	Person interface {
		GenericCRUD[model.Person]
	}

	// Application is application for a job offer
	Application interface {
		// Add creates new connection for applicant and job offer
		Add(ctx context.Context, jobID, applicantID string) (*model.Application, error)
	}
)

// NewSecurity creates job service
func NewSecurity(db *sql.DB) Security {
	return pgsvc.NewSecurity(db)
}

// NewJob creates job service
func NewJob(db *sql.DB) Job {
	return pgsvc.NewJob(db)
}

// NewPerson creates person service
func NewPerson(db *sql.DB) Person {
	return pgsvc.NewPerson(db)
}

// NewApplication creates application service
func NewApplication(db *sql.DB) Application {
	return pgsvc.NewApplication(db)
}

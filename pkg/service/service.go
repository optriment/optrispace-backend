package service

import (
	"context"
	"database/sql"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/pgsvc"
)

type (
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

	// Security creates user representation from echo.Context, if there is such data
	Security interface {
		// FromEchoContext acquires user from echo and persisten storage
		// It will construct *model.UserContext in the context too
		FromEchoContext(c echo.Context) (*model.UserContext, error)

		// FromLoginPassword creates UserContext from login and password in default realm
		FromLoginPassword(ctx context.Context, login, password string) (*model.UserContext, error)
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
		GenericCRUD[model.Application]
		// ListBy returns list of entities by specified filters
		// If jobID != "", method returns list of jobs
		ListBy(ctx context.Context, jobID, applicantID string) ([]*model.Application, error)
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

package service

import (
	"context"
	"database/sql"

	"github.com/labstack/echo/v4"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/ethsvc"
	"optrispace.com/work/pkg/service/pgsvc"
)

type (
	// GenericCRUD represents the standard methods for CRUD operations
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
		// FromEchoContext acquires user from echo and persistent storage
		// It will construct *model.UserContext in the context too
		FromEchoContext(c echo.Context) (*model.UserContext, error)

		// FromLoginPassword creates UserContext from login and password in default realm
		FromLoginPassword(ctx context.Context, login, password string) (*model.UserContext, error)
	}

	// Job handles job offers
	Job interface {
		// Add saves the entity into storage
		Add(ctx context.Context, customerID string, dto *model.CreateJobDTO) (*model.JobDTO, error)

		// Get returns a specific job by ID
		Get(ctx context.Context, id string) (*model.JobCardDTO, error)

		// List returns a list of jobs
		List(ctx context.Context) ([]*model.JobDTO, error)

		// Patch partially updates existing Job object
		Patch(ctx context.Context, id, customerID string, patch *model.UpdateJobDTO) (*model.JobDTO, error)

		// Block job
		Block(ctx context.Context, id, actorID string) error

		// Suspend job
		Suspend(ctx context.Context, id, actorID string) error

		// Resume job
		Resume(ctx context.Context, id, actorID string) error
	}

	// Person is a person who pay or earn
	Person interface {
		GenericCRUD[model.Person]

		// GetByAccessToken returns person by access token supplied
		GetByAccessToken(ctx context.Context, accessToken string) (*model.Person, error)

		// Update password
		UpdatePassword(ctx context.Context, subjectID, oldPassword, newPassword string) error

		// Patch partially updates existing Person object
		Patch(ctx context.Context, id, actorID string, patch map[string]any) (*model.BasicPersonDTO, error)

		// SetResources fully replace resources for the person
		SetResources(ctx context.Context, id, actorID string, resources []byte) error
	}

	// Application is application for a job offer
	Application interface {
		GenericCRUD[model.Application]
		// ListBy returns list of entities by specified filters
		// If jobID != "", method returns list of jobs
		// if actorID != "", method returns list of applications for job author or applications
		ListBy(ctx context.Context, jobID, actorID string) ([]*model.Application, error)

		// ListByApplicant returns list of applications by specified applicant
		ListByApplicant(ctx context.Context, applicantID string) ([]*model.Application, error)

		// GetChat returns chat associated with this application
		GetChat(ctx context.Context, id, actorID string) (*model.Chat, error)
	}

	// Contract is an agreement between a Customer and a Performer (Contractor)
	Contract interface {
		// Add saves the entity into storage
		Add(ctx context.Context, customerID string, dto *model.CreateContractDTO) (*model.ContractDTO, error)

		// Accept makes contract accepted by performer
		Accept(ctx context.Context, id, performerID string) (*model.ContractDTO, error)

		// Deploy makes contract deployed by customer
		Deploy(ctx context.Context, id, customerID string, contractParams *model.DeployContractDTO) (*model.ContractDTO, error)

		// Sign makes contract signed by performer
		Sign(ctx context.Context, id, performerID string) (*model.ContractDTO, error)

		// Fund makes contract funded by customer
		Fund(ctx context.Context, id, customerID string) (*model.ContractDTO, error)

		// Approve makes contract approved by customer
		Approve(ctx context.Context, id, customerID string) (*model.ContractDTO, error)

		// Complete makes contract completed by performer
		Complete(ctx context.Context, id, performerID string) (*model.ContractDTO, error)

		// GetByIDForPerson reads specified entity from storage by specified id and related for person
		// It can return model.ErrNotFound
		GetByIDForPerson(ctx context.Context, id, personID string) (*model.ContractDTO, error)

		// ListByPersonID returns list of entities by specific Person
		ListByPersonID(ctx context.Context, personID string) ([]*model.ContractDTO, error)
	}

	// Notification service manipulates with notifications
	Notification interface {
		// Push sends a data message to the configured channels (chats in Telegram messenger)
		Push(ctx context.Context, data string) error
	}

	// Stats service for statistic information
	Stats interface {
		// Stats returns users registrations number grouped by days
		Stats(ctx context.Context) (*model.Stats, error)
	}

	// Chat service for chat information, messaging etc.
	Chat interface {
		// AddMessage adds an message to the chat
		AddMessage(ctx context.Context, chatID, participantID, text string) (*model.Message, error)

		// Get returns chat description with all messages in it
		Get(ctx context.Context, chatID, participantID string) (*model.Chat, error)

		// ListByParticipant returns all chats by participant ID
		ListByParticipant(ctx context.Context, participantID string) ([]*model.ChatDTO, error)
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

// NewContract creates contract service
func NewContract(db *sql.DB, eth ethsvc.Ethereum) Contract {
	return pgsvc.NewContract(db, eth)
}

// NewNotification creates notification service
func NewNotification(tgToken string, chatIDs ...int64) Notification {
	return pgsvc.NewNotification(tgToken, chatIDs...)
}

// NewStats create stats service
func NewStats(db *sql.DB) Stats {
	return pgsvc.NewStats(db)
}

// NewChat create chat service
func NewChat(db *sql.DB) Chat {
	return pgsvc.NewChat(db)
}

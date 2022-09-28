package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Main (main!) purposes of these types is a representation
// entities when returning to the REST requester
// Therefore there are computed fields in these types
type (
	// Job is a job offer publication
	Job struct {
		ID                string          `json:"id"`
		Title             string          `json:"title"`
		Description       string          `json:"description"`
		Budget            decimal.Decimal `json:"budget"`
		Duration          int32           `json:"duration,omitempty"`
		CreatedAt         time.Time       `json:"created_at"`
		UpdatedAt         time.Time       `json:"updated_at"`
		CreatedBy         string          `json:"created_by"`
		ApplicationsCount uint            `json:"applications_count"`
		Customer          *Person         `json:"customer,omitempty"`
	}

	// Person — customer, executor, seller, buyer etc.
	Person struct {
		ID              string    `json:"id"`
		Realm           string    `json:"realm"`
		Login           string    `json:"login"`
		Password        string    `json:"password,omitempty"`
		DisplayName     string    `json:"display_name"`
		CreatedAt       time.Time `json:"created_at"`
		Email           string    `json:"email"`
		EthereumAddress string    `json:"ethereum_address"`
		Resources       string    `json:"resources"`
		AccessToken     string    `json:"-"`
		IsAdmin         bool      `json:"is_admin"`
	}

	// Contract is a contract for execution some a task and
	// a payment obligation
	Contract struct {
		ID               string          `json:"id,omitempty"`
		Customer         *Person         `json:"customer,omitempty"`
		Performer        *Person         `json:"performer,omitempty"`
		Application      *Application    `json:"application,omitempty"`
		Title            string          `json:"title,omitempty"`
		Description      string          `json:"description,omitempty"`
		Price            decimal.Decimal `json:"price,omitempty"`
		Duration         int32           `json:"duration,omitempty"`
		Status           string          `json:"status,omitempty"`
		CreatedAt        time.Time       `json:"created_at,omitempty"`
		UpdatedAt        time.Time       `json:"updated_at,omitempty"`
		CreatedBy        string          `json:"created_by,omitempty"`
		ContractAddress  string          `json:"contract_address,omitempty"`
		CustomerAddress  string          `json:"customer_address,omitempty"`
		PerformerAddress string          `json:"performer_address,omitempty"`
	}

	// ContractStatus represents a contract status
	ContractStatus string

	// Application is an application for a job
	Application struct {
		ID        string          `json:"id,omitempty"`
		CreatedAt time.Time       `json:"created_at,omitempty"`
		UpdatedAt time.Time       `json:"updated_at,omitempty"`
		Applicant *JobApplicant   `json:"applicant,omitempty"`
		Comment   string          `json:"comment,omitempty"`
		Price     decimal.Decimal `json:"price,omitempty"`
		Job       *Job            `json:"job,omitempty"`
		Contract  *Contract       `json:"contract,omitempty"`
	}

	// Stats for statistics
	Stats struct {
		Registrations map[string]int `json:"registrations,omitempty"`
	}

	// Chat is a chat instance
	Chat struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		Topic     string    `json:"topic"`
		Messages  []Message `json:"messages,omitempty"`
	}

	// Message is a message instance
	Message struct {
		ID         string    `json:"id"`
		ChatID     string    `json:"chat_id"`
		CreatedAt  time.Time `json:"created_at"`
		CreatedBy  string    `json:"created_by"`
		Text       string    `json:"text"`
		AuthorName string    `json:"author_name"`
	}

	// JobApplicant represents a person who applied for specific job
	JobApplicant struct {
		ID              string `json:"id"`
		DisplayName     string `json:"display_name"`
		EthereumAddress string `json:"ethereum_address"`
		Resources       string `json:"resources"`
	}
)

// Contract statuses
const (
	ContractCreated   = "created"
	ContractAccepted  = "accepted"
	ContractDeployed  = "deployed"
	ContractSigned    = "signed"
	ContractFunded    = "funded"
	ContractApproved  = "approved"
	ContractCompleted = "completed"
)

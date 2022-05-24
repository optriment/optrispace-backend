package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Main (main!) purposes of these types is a representation
// entities when returning to the REST requester
// Thererefore there are computed fields in these types
type (
	// Job is a job offer publication
	Job struct {
		ID                string          `json:"id"`
		Title             string          `json:"title"`
		Description       string          `json:"description"`
		Budget            decimal.Decimal `json:"budget"`
		Duration          int32           `json:"duration"`
		CreatedAt         time.Time       `json:"created_at"`
		UpdatedAt         time.Time       `json:"updated_at"`
		CreatedBy         string          `json:"created_by"`
		ApplicationsCount uint            `json:"applications_count"`
		Customer          *Person         `json:"customer"`
	}

	// Person — customer, executor, seller, buyer etc.
	Person struct {
		ID              string    `json:"id,omitempty"`
		Realm           string    `json:"realm,omitempty"`
		Login           string    `json:"login,omitempty"`
		Password        string    `json:"password,omitempty"`
		DisplayName     string    `json:"display_name,omitempty"`
		CreatedAt       time.Time `json:"created_at,omitempty"`
		Email           string    `json:"email,omitempty"`
		EthereumAddress string    `json:"ethereum_address,omitempty"`
	}

	// Project is a sequence of contracts group
	Project struct{}

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
		Applicant *Person         `json:"applicant,omitempty"`
		Comment   string          `json:"comment,omitempty"`
		Price     decimal.Decimal `json:"price,omitempty"`
		Job       *Job            `json:"job,omitempty"`
		Contract  *Contract       `json:"contract,omitempty"`
	}
)

// Contract statuses
const (
	ContractCreated  = "created"
	ContractAccepted = "accepted"
	ContractDeployed = "deployed"
	ContractSent     = "sent"
	ContractApproved = "approved"
)

// var allContractStatus = []string{ContractCreated, ContractAccepted, ContractSent, ContractApproved}

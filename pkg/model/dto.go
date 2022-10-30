package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	// CreateJobDTO is a contract representation on creation process
	CreateJobDTO struct {
		Title       string `validate:"required"`
		Description string `validate:"required"`
		Budget      decimal.Decimal
		Duration    int32
	}

	// JobDTO is a representation of the job
	JobDTO struct {
		ID                      string          `json:"id"`
		Title                   string          `json:"title"`
		Description             string          `json:"description"`
		Budget                  decimal.Decimal `json:"budget"`
		Duration                int32           `json:"duration,omitempty"`
		CreatedAt               time.Time       `json:"created_at"`
		UpdatedAt               time.Time       `json:"updated_at"`
		CreatedBy               string          `json:"created_by"`
		ApplicationsCount       uint            `json:"applications_count"`
		CustomerDisplayName     string          `json:"customer_display_name"`
		CustomerEthereumAddress string          `json:"customer_ethereum_address"`
	}

	// JobCardDTO is a representation of the job with extended attributes
	JobCardDTO struct {
		JobDTO

		IsSuspended bool `json:"is_suspended"`
	}

	// UpdateJobDTO is a job representation on updation process
	UpdateJobDTO struct {
		Title       string `validate:"required"`
		Description string `validate:"required"`
		Budget      decimal.Decimal
		Duration    int32
	}

	// CreateContractDTO is a contract representation on creation process
	CreateContractDTO struct {
		ApplicationID string          `validate:"required"`
		Title         string          `validate:"required"`
		Description   string          `validate:"required"`
		Price         decimal.Decimal `validate:"required"`
		Duration      int32
	}

	// DeployContractDTO is a contract representation on deployment process
	DeployContractDTO struct {
		ContractAddress string `validate:"required"`
	}

	// ContractDTO is a representation of contract
	ContractDTO struct {
		ID                   string          `json:"id"`
		CustomerID           string          `json:"customer_id"`
		PerformerID          string          `json:"performer_id"`
		ApplicationID        string          `json:"application_id"`
		CustomerDisplayName  string          `json:"customer_display_name"`
		PerformerDisplayName string          `json:"performer_display_name"`
		Title                string          `json:"title"`
		Description          string          `json:"description"`
		Price                decimal.Decimal `json:"price"`
		Duration             int32           `json:"duration"`
		Status               string          `json:"status"`
		CreatedAt            time.Time       `json:"created_at"`
		UpdatedAt            time.Time       `json:"updated_at"`
		CreatedBy            string          `json:"created_by"`
		ContractAddress      string          `json:"contract_address"`
		CustomerAddress      string          `json:"customer_address"`
		PerformerAddress     string          `json:"performer_address"`
	}

	// BasicPersonDTO is a representation of a person excepts restricted fields
	BasicPersonDTO struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Email           string `json:"email"`
		EthereumAddress string `json:"ethereum_address"`
		Resources       string `json:"resources"`
	}

	// ChatDTO represents basic information about chat
	ChatDTO struct {
		ID            string            `json:"id"`
		Topic         string            `json:"topic"`
		Kind          string            `json:"kind"`
		Title         string            `json:"title"`
		JobID         string            `json:"job_id,omitempty"`
		ApplicationID string            `json:"application_id,omitempty"`
		ContractID    string            `json:"contract_id,omitempty"`
		Participants  []*ParticipantDTO `json:"participants,omitempty"`
	}

	// ParticipantDTO represents is a chat participant
	ParticipantDTO struct {
		ID              string `json:"id"`
		DisplayName     string `json:"display_name"`
		EthereumAddress string `json:"ethereum_address"`
	}

	// CreateApplicationDTO is an application representation on applyment process
	CreateApplicationDTO struct {
		JobID   string `validate:"required"`
		Comment string `validate:"required"`
		Price   decimal.Decimal
	}

	// ApplicationDTO is an application for a job
	ApplicationDTO struct {
		ID                       string          `json:"id"`
		JobID                    string          `json:"job_id"`
		JobTitle                 string          `json:"job_title"`
		JobDescription           string          `json:"job_description"`
		JobBudget                decimal.Decimal `json:"job_budget"`
		ApplicantID              string          `json:"applicant_id"`
		ContractID               string          `json:"contract_id"`
		ContractStatus           string          `json:"contract_status"`
		Comment                  string          `json:"comment"`
		Price                    decimal.Decimal `json:"price"`
		ApplicantEthereumAddress string          `json:"applicant_ethereum_address"`
		ApplicantDisplayName     string          `json:"applicant_display_name"`
		CreatedAt                time.Time       `json:"created_at"`
	}
)

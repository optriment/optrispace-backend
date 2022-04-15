package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	// Job is a job offer publication
	Job struct {
		ID                string          `json:"id,omitempty"`
		Title             string          `json:"title,omitempty"`
		Description       string          `json:"description,omitempty"`
		Budget            decimal.Decimal `json:"budget,omitempty"`
		Duration          int32           `json:"duration,omitempty"`
		CreatedAt         time.Time       `json:"created_at,omitempty"`
		CreatedBy         *Person         `json:"created_by,omitempty"`
		UpdatedAt         time.Time       `json:"updated_at,omitempty"`
		ApplicationsCount uint            `json:"applications_count,omitempty"`
	}

	// Person — customer, executor, seller, buyer etc.
	Person struct {
		ID        string    `json:"id,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		Address   string    `json:"address,omitempty"`
	}

	// Project is a sequence of contracts group
	Project struct{}

	// Contract is a contract for execution some a task and
	// a payment obligation
	Contract struct{}

	// Application is an application for a job
	Application struct {
		ID        string    `json:"id,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		Applicant *Person   `json:"applicant,omitempty"`
		Job       *Job      `json:"job,omitempty"`
	}
)

package model

import "time"

type (
	// Job is a job offer publication
	Job struct {
		ID           string    `json:"id,omitempty"`
		CreationTime time.Time `json:"creationTime,omitempty"`
		Title        string    `json:"title,omitempty"`
		Description  string    `json:"description,omitempty"`
		Customer     *Person   `json:"customer,omitempty"`
	}

	// Person: customer, executor, seller, buyer etc.
	Person struct {
		ID      string `json:"id,omitempty"`
		Address string `json:"address,omitempty"`
	}

	// Project is a sequence of contracts group
	Project struct{}

	// Contract is a contract for execution some a task and
	// a payment obligation
	Contract struct{}

	// Application is an application for a job
	Application struct {
		ID           string    `json:"id,omitempty"`
		CreationTime time.Time `json:"creationTime,omitempty"`
		Applicant    *Person   `json:"applicant,omitempty"`
		Job          *Job      `json:"job,omitempty"`
	}
)

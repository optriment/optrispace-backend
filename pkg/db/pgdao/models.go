// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0

package pgdao

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Applications for job offers
type Application struct {
	// PK
	ID string
	// Application timestamp. When application was created.
	CreatedAt time.Time
	// Application update timestamp. When application was updated last time.
	UpdatedAt time.Time
	// Applicant's initial comment on the application
	Comment string
	// Proposed price
	Price string
	// Job offer
	JobID string
	// Potential performer
	ApplicantID string
}

// Contracts table
type Contract struct {
	// PK
	ID string
	// Customer for the job. Who paying.
	CustomerID string
	// Person who performing the job
	PerformerID string
	// Application was created before the contract
	ApplicationID string
	// Contract title. Like "web site creation". Can be copied from the appropriate job.
	Title string
	// Details about the contract. May be long, long text. Also can be copied from the appropriate job.
	Description string
	// The crontract price
	Price string
	// The contract duration
	Duration sql.NullInt32
	// Current status of the contract
	Status    string
	CreatedBy string
	// Creation timestamp
	CreatedAt time.Time
	// When the contract was updated last time
	UpdatedAt time.Time
	// Customer address in in block chain
	CustomerAddress string
	// Performer address in in block chain
	PerformerAddress string
	// Address in the block chain relevant smart contract
	ContractAddress string
}

// Job offer table
type Job struct {
	// PK
	ID string
	// Job title. Like "web site creation"
	Title string
	// Details about the job. May be long, long text
	Description string
	// Estimated cost of the job if any
	Budget sql.NullString
	// Estimated duration of the job in days if any
	Duration sql.NullInt32
	// Creation timestamp
	CreatedAt time.Time
	// When the job was updated last time
	UpdatedAt time.Time
	// Who created this job and should pay for it
	CreatedBy string
}

// Person who can pay, get or earn money
type Person struct {
	// PK
	ID string
	// Authentication realm (inhouse by default)
	Realm string
	// Login for authentication (must be unique for the separate authentication realm)
	Login string
	// Salty password hash
	PasswordHash string
	// User name for displaying
	DisplayName string
	// Creation time
	CreatedAt time.Time
	// User Email
	Email string
	// Person address in Ethereum-compatible block chains
	EthereumAddress string
	// Person's resources list. They may be links to social networks, portfolio, messenger IDs etc
	Resources json.RawMessage
	// Person's personal access token for Bearer authentication schema
	AccessToken sql.NullString
	// Does user have admin privileges?
	IsAdmin bool
}

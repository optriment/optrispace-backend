package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// ContractSvc is a contract service
	ContractSvc struct {
		db *sql.DB
	}
)

// NewContract creates service
func NewContract(db *sql.DB) *ContractSvc {
	return &ContractSvc{db: db}
}

// Add implements Contract interface
func (s *ContractSvc) Add(ctx context.Context, contract *model.Contract) (*model.Contract, error) {
	var result *model.Contract
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		id := pgdao.NewID()

		contractParams := pgdao.ContractAddParams{
			ID:            id,
			CustomerID:    contract.Customer.ID,
			PerformerID:   contract.Performer.ID,
			ApplicationID: contract.Application.ID,
			Title:         contract.Title,
			Description:   contract.Description,
			Price:         contract.Price.String(),
			Duration: sql.NullInt32{
				Int32: contract.Duration,
				Valid: contract.Duration > 0,
			},
			CreatedBy: contract.CreatedBy,
		}

		newContract, err := queries.ContractAdd(ctx, contractParams)

		if pqe, ok := err.(*pq.Error); ok { //nolint: errorlint
			if pqe.Code == "23505" {
				return fmt.Errorf("%s: %w", pqe.Detail, model.ErrDuplication)
			}
		}

		if err != nil {
			return fmt.Errorf("unable to ContractAdd: %w", err)
		}

		result = &model.Contract{
			ID:          newContract.ID,
			Customer:    &model.Person{ID: newContract.CustomerID},
			Performer:   &model.Person{ID: newContract.PerformerID},
			Application: &model.Application{ID: newContract.ApplicationID},
			Title:       newContract.Title,
			Description: newContract.Description,
			Price:       decimal.RequireFromString(newContract.Price),
			Duration:    newContract.Duration.Int32,
			CreatedAt:   newContract.CreatedAt,
			UpdatedAt:   newContract.UpdatedAt,
		}

		return nil
	})
}

// GetByIDForPerson loads contract by ID related for specific person
func (s *ContractSvc) GetByIDForPerson(ctx context.Context, id, personID string) (*model.Contract, error) {
	var result *model.Contract
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		contractParams := pgdao.ContractGetByIDAndPersonIDParams{
			ID:       id,
			PersonID: personID,
		}

		a, err := queries.ContractGetByIDAndPersonID(ctx, contractParams)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("unable to ContractGet with id=%s: %w", id, err)
		}

		result = &model.Contract{
			ID:          a.ID,
			CreatedAt:   a.CreatedAt,
			UpdatedAt:   a.UpdatedAt,
			Customer:    &model.Person{ID: a.CustomerID},
			Performer:   &model.Person{ID: a.PerformerID},
			Title:       a.Title,
			Description: a.Description,
			Price:       decimal.RequireFromString(a.Price),
			Duration:    a.Duration.Int32,
			Application: &model.Application{ID: a.ApplicationID},
			CreatedBy:   a.CreatedBy,
		}

		return nil
	})
}

// ListByPersonID loads all related contracts for specific person
func (s *ContractSvc) ListByPersonID(ctx context.Context, personID string) ([]*model.Contract, error) {
	var result []*model.Contract = make([]*model.Contract, 0)

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		aa, err := queries.ContractsGetByPerson(ctx, personID)
		if err != nil {
			return fmt.Errorf("unable to ContractsListBy: %w", err)
		}

		for _, a := range aa {
			result = append(result, &model.Contract{
				ID:          a.ID,
				CreatedAt:   a.CreatedAt,
				UpdatedAt:   a.UpdatedAt,
				Customer:    &model.Person{ID: a.CustomerID},
				Performer:   &model.Person{ID: a.PerformerID},
				Title:       a.Title,
				Description: a.Description,
				Price:       decimal.RequireFromString(a.Price),
				Duration:    a.Duration.Int32,
			})
		}

		return nil
	})
}

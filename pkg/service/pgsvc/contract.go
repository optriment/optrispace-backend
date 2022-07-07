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
			CreatedBy:       contract.CreatedBy,
			CustomerAddress: contract.CustomerAddress,
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

		result = &model.Contract{ //nolint: dupl
			ID:               newContract.ID,
			Customer:         &model.Person{ID: newContract.CustomerID},
			Performer:        &model.Person{ID: newContract.PerformerID},
			Application:      &model.Application{ID: newContract.ApplicationID},
			Title:            newContract.Title,
			Description:      newContract.Description,
			Price:            decimal.RequireFromString(newContract.Price),
			Duration:         newContract.Duration.Int32,
			Status:           newContract.Status,
			CreatedAt:        newContract.CreatedAt,
			UpdatedAt:        newContract.UpdatedAt,
			CreatedBy:        newContract.CreatedBy,
			ContractAddress:  newContract.ContractAddress,
			CustomerAddress:  newContract.CustomerAddress,
			PerformerAddress: newContract.PerformerAddress,
		}

		return nil
	})
}

func contractByIDPersonID(ctx context.Context, queries *pgdao.Queries, id, personID string) (*model.Contract, error) {
	contractParams := pgdao.ContractGetByIDAndPersonIDParams{
		ID:       id,
		PersonID: personID,
	}

	a, err := queries.ContractGetByIDAndPersonID(ctx, contractParams)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrEntityNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("unable to ContractGet with id=%s: %w", id, err)
	}

	return &model.Contract{ //nolint: dupl
		ID:               a.ID,
		Customer:         &model.Person{ID: a.CustomerID},
		Performer:        &model.Person{ID: a.PerformerID},
		Application:      &model.Application{ID: a.ApplicationID},
		Title:            a.Title,
		Description:      a.Description,
		Price:            decimal.RequireFromString(a.Price),
		Duration:         a.Duration.Int32,
		Status:           a.Status,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
		CreatedBy:        a.CreatedBy,
		ContractAddress:  a.ContractAddress,
		CustomerAddress:  a.CustomerAddress,
		PerformerAddress: a.PerformerAddress,
	}, nil
}

// GetByIDForPerson loads contract by ID related for specific person
func (s *ContractSvc) GetByIDForPerson(ctx context.Context, id, personID string) (*model.Contract, error) {
	var result *model.Contract
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		r, err := contractByIDPersonID(ctx, queries, id, personID)
		result = r
		return err
	})
}

// ListByPersonID loads all related contracts for specific person
func (s *ContractSvc) ListByPersonID(ctx context.Context, personID string) ([]*model.Contract, error) {
	result := make([]*model.Contract, 0)

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		aa, err := queries.ContractsGetByPerson(ctx, personID)
		if err != nil {
			return fmt.Errorf("unable to ContractsListBy: %w", err)
		}

		for _, a := range aa {
			result = append(result, &model.Contract{
				ID:          a.ID,
				Customer:    &model.Person{ID: a.CustomerID},
				Performer:   &model.Person{ID: a.PerformerID},
				Application: &model.Application{},
				Title:       a.Title,
				Description: a.Description,
				Price:       decimal.RequireFromString(a.Price),
				Duration:    a.Duration.Int32,
				Status:      a.Status,
				CreatedAt:   a.CreatedAt,
				UpdatedAt:   a.UpdatedAt,
			})
		}

		return nil
	})
}

func (s *ContractSvc) toStatus(ctx context.Context, actorID string, patchParams *pgdao.ContractPatchParams, validator func(c *model.Contract) error) (*model.Contract, error) {
	var result *model.Contract

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		c, err := contractByIDPersonID(ctx, queries, patchParams.ID, actorID)
		if err != nil {
			return err
		}

		if e := validator(c); e != nil {
			return e
		}

		o, err := queries.ContractPatch(ctx, *patchParams)

		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		result = &model.Contract{ //nolint: dupl
			ID:               o.ID,
			Title:            o.Title,
			Description:      o.Description,
			Price:            decimal.RequireFromString(o.Price),
			Duration:         o.Duration.Int32,
			Status:           o.Status,
			CreatedAt:        o.CreatedAt,
			UpdatedAt:        o.UpdatedAt,
			CreatedBy:        o.CreatedBy,
			ContractAddress:  o.ContractAddress,
			CustomerAddress:  o.CustomerAddress,
			PerformerAddress: o.PerformerAddress,
		}

		return err
	})
}

// Accept makes contract accepted if any
func (s *ContractSvc) Accept(ctx context.Context, id, actorID, performerAddress string) (*model.Contract, error) {
	allowedSourceStatus := model.ContractCreated
	targetStatus := model.ContractAccepted
	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange:           true,
		Status:                 targetStatus,
		PerformerAddressChange: true,
		PerformerAddress:       performerAddress,
		ID:                     id,
	}, func(c *model.Contract) error {
		if c.Performer.ID != actorID {
			return model.ErrInsufficientRights
		}
		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}
		return nil
	})
}

// Deploy makes contract deployed (in the target blockchain) if any
func (s *ContractSvc) Deploy(ctx context.Context, id, actorID, contractAddress string) (*model.Contract, error) {
	allowedSourceStatus := model.ContractAccepted
	targetStatus := model.ContractDeployed
	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange:          true,
		Status:                targetStatus,
		ContractAddressChange: true,
		ContractAddress:       contractAddress,
		ID:                    id,
	}, func(c *model.Contract) error {
		if c.Customer.ID != actorID {
			return model.ErrInsufficientRights
		}
		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}
		return nil
	})
}

// Send makes contract sent if any
func (s *ContractSvc) Send(ctx context.Context, id, actorID string) (*model.Contract, error) {
	allowedSourceStatus := model.ContractDeployed
	targetStatus := model.ContractSent
	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.Contract) error {
		if c.Performer.ID != actorID {
			return model.ErrInsufficientRights
		}
		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}
		return nil
	})
}

// Approve makes contract approved if any
func (s *ContractSvc) Approve(ctx context.Context, id, actorID string) (*model.Contract, error) {
	allowedSourceStatus := model.ContractSent
	targetStatus := model.ContractApproved
	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.Contract) error {
		if c.Customer.ID != actorID {
			return model.ErrInsufficientRights
		}
		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}
		return nil
	})
}

// Complete makes contract completed if any
func (s *ContractSvc) Complete(ctx context.Context, id, actorID string) (*model.Contract, error) {
	allowedSourceStatus := model.ContractApproved
	targetStatus := model.ContractCompleted
	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.Contract) error {
		if c.Performer.ID != actorID {
			return model.ErrInsufficientRights
		}
		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}
		return nil
	})
}

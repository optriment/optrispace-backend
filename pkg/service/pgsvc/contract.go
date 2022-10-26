package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service/ethsvc"
)

type (
	// ContractSvc is a contract service
	ContractSvc struct {
		db  *sql.DB
		eth ethsvc.Ethereum
	}
)

// NewContract creates service
func NewContract(db *sql.DB, eth ethsvc.Ethereum) *ContractSvc {
	return &ContractSvc{
		db:  db,
		eth: eth,
	}
}

// Add implements Contract interface
func (s *ContractSvc) Add(ctx context.Context, customerID string, dto *model.CreateContractDTO) (*model.ContractDTO, error) {
	var result *model.ContractDTO

	if strings.TrimSpace(dto.ApplicationID) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("application_id"),
		}
	}

	if strings.TrimSpace(dto.Title) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("title"),
		}
	}

	if strings.TrimSpace(dto.Description) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("description"),
		}
	}

	if decimal.Zero.Equal(dto.Price) {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("price"),
		}
	}

	if dto.Price.IsNegative() {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustBePositive("price"),
		}
	}

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		application, err := queries.ApplicationGet(ctx, strings.TrimSpace(dto.ApplicationID))
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "application does not exist",
			}
		}

		customer, err := queries.PersonGet(ctx, customerID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "customer does not exist",
			}
		}

		customerEthereumAddress := strings.ToLower(strings.TrimSpace(customer.EthereumAddress))
		if customerEthereumAddress == "" {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "customer does not have wallet",
			}
		}

		performer, err := queries.PersonGet(ctx, application.ApplicantID)
		if err != nil {
			return &model.BackendError{
				Cause:   model.ErrEntityNotFound,
				Message: "performer does not exist",
			}
		}

		performerEthereumAddress := strings.ToLower(strings.TrimSpace(performer.EthereumAddress))
		if performerEthereumAddress == "" {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "performer does not have wallet",
			}
		}

		if strings.EqualFold(performerEthereumAddress, customerEthereumAddress) {
			return &model.BackendError{
				Cause:   model.ErrValidationFailed,
				Message: "customer and performer addresses cannot be the same",
			}
		}

		contractParams := pgdao.ContractAddParams{
			ID:            pgdao.NewID(),
			CustomerID:    customer.ID,
			PerformerID:   application.ApplicantID,
			ApplicationID: strings.TrimSpace(application.ID),
			Title:         strings.TrimSpace(dto.Title),
			Description:   strings.TrimSpace(dto.Description),
			Price:         dto.Price.String(),
			Duration: sql.NullInt32{
				Int32: dto.Duration,
				Valid: dto.Duration > 0,
			},
			CreatedBy:        customer.ID,
			Status:           model.ContractCreated,
			CustomerAddress:  customerEthereumAddress,
			PerformerAddress: performerEthereumAddress,
		}

		newContract, err := queries.ContractAdd(ctx, contractParams)

		if pqe, ok := err.(*pq.Error); ok { //nolint: errorlint
			if pqe.Code == "23505" {
				return &model.BackendError{
					Cause:   model.ErrDuplication,
					Message: "contract already exists",
				}
			}
		}

		if err != nil {
			return fmt.Errorf("unable to ContractAdd: %w", err)
		}

		result = &model.ContractDTO{ //nolint: dupl
			ID:               newContract.ID,
			CustomerID:       newContract.CustomerID,
			PerformerID:      newContract.PerformerID,
			ApplicationID:    newContract.ApplicationID,
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

// Accept makes contract accepted
func (s *ContractSvc) Accept(ctx context.Context, id, actorID string) (*model.ContractDTO, error) {
	allowedSourceStatus := model.ContractCreated
	targetStatus := model.ContractAccepted

	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.ContractDTO) error {
		if c.PerformerID != actorID {
			return model.ErrInsufficientRights
		}

		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}

		return nil
	})
}

// Deploy makes contract deployed
func (s *ContractSvc) Deploy(ctx context.Context, id, actorID string, dto *model.DeployContractDTO) (*model.ContractDTO, error) {
	allowedSourceStatus := model.ContractAccepted
	targetStatus := model.ContractDeployed

	contractAddress := strings.ToLower(strings.TrimSpace(dto.ContractAddress))
	if contractAddress == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("contract_address"),
		}
	}

	if !common.IsHexAddress(contractAddress) {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorInvalidFormat("contract_address"),
		}
	}

	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange:          true,
		Status:                targetStatus,
		ContractAddressChange: true,
		ContractAddress:       contractAddress,
		ID:                    id,
	}, func(c *model.ContractDTO) error {
		if c.CustomerID != actorID {
			return model.ErrInsufficientRights
		}

		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}

		return nil
	})
}

// Sign makes contract signed
func (s *ContractSvc) Sign(ctx context.Context, id, actorID string) (*model.ContractDTO, error) {
	allowedSourceStatus := model.ContractDeployed
	targetStatus := model.ContractSigned

	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.ContractDTO) error {
		if c.PerformerID != actorID {
			return model.ErrInsufficientRights
		}

		if !common.IsHexAddress(c.ContractAddress) {
			return &model.BackendError{
				Cause:    model.ErrValidationFailed,
				Message:  model.ValidationErrorInvalidFormat("contract_address"),
				TechInfo: c.ContractAddress,
			}
		}

		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}

		return nil
	})
}

// Fund makes contract funded
func (s *ContractSvc) Fund(ctx context.Context, id, actorID string) (*model.ContractDTO, error) {
	allowedSourceStatus := model.ContractSigned
	targetStatus := model.ContractFunded

	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.ContractDTO) error {
		if c.CustomerID != actorID {
			return model.ErrInsufficientRights
		}

		if !common.IsHexAddress(c.ContractAddress) {
			return &model.BackendError{
				Cause:    model.ErrValidationFailed,
				Message:  model.ValidationErrorInvalidFormat("contract_address"),
				TechInfo: c.ContractAddress,
			}
		}

		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}

		return s.checkAddressBalance(ctx, c.Price, c.ContractAddress)
	})
}

// Approve makes contract approved
func (s *ContractSvc) Approve(ctx context.Context, id, actorID string) (*model.ContractDTO, error) {
	allowedSourceStatus := model.ContractFunded
	targetStatus := model.ContractApproved

	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.ContractDTO) error {
		if c.CustomerID != actorID {
			return model.ErrInsufficientRights
		}

		if !common.IsHexAddress(c.ContractAddress) {
			return &model.BackendError{
				Cause:    model.ErrValidationFailed,
				Message:  model.ValidationErrorInvalidFormat("contract_address"),
				TechInfo: c.ContractAddress,
			}
		}

		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}

		return nil
	})
}

// Complete makes contract completed
func (s *ContractSvc) Complete(ctx context.Context, id, actorID string) (*model.ContractDTO, error) {
	allowedSourceStatus := model.ContractApproved
	targetStatus := model.ContractCompleted

	return s.toStatus(ctx, actorID, &pgdao.ContractPatchParams{
		StatusChange: true,
		Status:       targetStatus,
		ID:           id,
	}, func(c *model.ContractDTO) error {
		if c.PerformerID != actorID {
			return model.ErrInsufficientRights
		}

		if !common.IsHexAddress(c.ContractAddress) {
			return &model.BackendError{
				Cause:    model.ErrValidationFailed,
				Message:  model.ValidationErrorInvalidFormat("contract_address"),
				TechInfo: c.ContractAddress,
			}
		}

		if c.Status != allowedSourceStatus {
			return fmt.Errorf("%w: unable to move from %s to %s", model.ErrInappropriateAction, c.Status, targetStatus)
		}

		return nil
	})
}

// GetByIDForPerson loads contract by ID related for specific person
func (s *ContractSvc) GetByIDForPerson(ctx context.Context, id, personID string) (*model.ContractDTO, error) {
	var result *model.ContractDTO
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		r, err := contractByIDPersonID(ctx, queries, id, personID)
		result = r
		return err
	})
}

// ListByPersonID loads all related contracts for specific person
func (s *ContractSvc) ListByPersonID(ctx context.Context, personID string) ([]*model.ContractDTO, error) {
	result := make([]*model.ContractDTO, 0)

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		aa, err := queries.ContractsGetByPerson(ctx, personID)
		if err != nil {
			return fmt.Errorf("unable to ContractsGetByPerson: %w", err)
		}

		for _, a := range aa {
			result = append(result, &model.ContractDTO{
				ID:                   a.ID,
				CustomerID:           a.CustomerID,
				PerformerID:          a.PerformerID,
				ApplicationID:        a.ApplicationID,
				CustomerDisplayName:  a.CustomerName.String,
				PerformerDisplayName: a.PerformerName.String,
				ContractAddress:      a.ContractAddress,
				CustomerAddress:      a.CustomerAddress,
				PerformerAddress:     a.PerformerAddress,
				Title:                a.Title,
				Description:          a.Description,
				Price:                decimal.RequireFromString(a.Price),
				Duration:             a.Duration.Int32,
				Status:               a.Status,
				CreatedBy:            a.CreatedBy,
				CreatedAt:            a.CreatedAt,
				UpdatedAt:            a.UpdatedAt,
			})
		}

		return nil
	})
}

// checkAddressBalance checks that contract have enough coins to supply contract entity
// It should return nil if there are enough money at the contract address in the chain
func (s *ContractSvc) checkAddressBalance(ctx context.Context, requiredBalance decimal.Decimal, contractAddress string) error {
	// NOTE: If you have an issue with getting balance from blockchain by contract address,
	// please try to choose another server from https://chainlist.org/chain/97 and update ./testdata/dev.yaml
	balance, err := s.eth.Balance(ctx, contractAddress)
	if err != nil {
		return err
	}

	if balance.LessThan(requiredBalance) {
		return &model.BackendError{
			Cause:    model.ErrInsufficientFunds,
			Message:  "the contract does not have sufficient funds",
			TechInfo: contractAddress,
		}
	}

	return nil
}

func (s *ContractSvc) toStatus(ctx context.Context, actorID string, patchParams *pgdao.ContractPatchParams, validator func(c *model.ContractDTO) error) (*model.ContractDTO, error) {
	var result *model.ContractDTO

	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		c, err := contractByIDPersonID(ctx, queries, patchParams.ID, actorID)
		if err != nil {
			return err
		}

		if e := validator(c); e != nil {
			return e
		}

		o, err := queries.ContractPatch(ctx, *patchParams)
		if err != nil {
			return fmt.Errorf("unable to ContractPatch with id=%s: %w", patchParams.ID, err)
		}

		result = &model.ContractDTO{ //nolint: dupl
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

func contractByIDPersonID(ctx context.Context, queries *pgdao.Queries, id, personID string) (*model.ContractDTO, error) {
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

	return &model.ContractDTO{ //nolint: dupl
		ID:               a.ID,
		CustomerID:       a.CustomerID,
		PerformerID:      a.PerformerID,
		ApplicationID:    a.ApplicationID,
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

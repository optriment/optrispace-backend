package controller

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	// Contract controller
	Contract struct {
		sm  service.Security
		svc service.Contract
	}
)

func validateStruct(s interface{}) error {
	validate := validator.New()

	// Register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	err := validate.Struct(s)

	if err == nil {
		return nil
	}

	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		for _, err := range ve {
			if err.Tag() == "required" {
				return &model.BackendError{
					Cause:   model.ErrValidationFailed,
					Message: model.ValidationErrorRequired(err.Field()),
				}
			}
		}

		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: ve.Error(),
		}
	}

	return &model.BackendError{
		Cause:   model.ErrValidationFailed,
		Message: err.Error(),
	}
}

// NewContract create new service
func NewContract(sm service.Security, svc service.Contract) Registerer {
	return &Contract{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Contract) Register(e *echo.Echo) {
	e.POST(resourceContract, cont.add)
	e.GET(resourceContract, cont.list)
	e.GET(resourceContract+"/:id", cont.get)
	e.POST(resourceContract+"/:id/accept", cont.accept)
	e.POST(resourceContract+"/:id/deploy", cont.deploy)
	e.POST(resourceContract+"/:id/sign", cont.sign)
	e.POST(resourceContract+"/:id/fund", cont.fund)
	e.POST(resourceContract+"/:id/approve", cont.approve)
	e.POST(resourceContract+"/:id/complete", cont.complete)
	log.Debug().Str("controller", resourceContract).Msg("Registered")
}

type createContractParams struct {
	ApplicationID string          `json:"application_id" validate:"required"`
	Title         string          `json:"title" validate:"required"`
	Description   string          `json:"description" validate:"required"`
	Price         decimal.Decimal `json:"price" validate:"required"`
	Duration      int32           `json:"duration"`
}

// @Summary     Create a new contract
// @Description Creates a new contract based on existent application.
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       job body     controller.createContractParams true "Contract Params"
// @Success     201 {object} model.ContractDTO
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     409 {object} model.BackendError "duplication"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts [post]
func (cont *Contract) add(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(createContractParams)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if err = validateStruct(ie); err != nil {
		return err
	}

	dto := model.CreateContractDTO{
		ApplicationID: ie.ApplicationID,
		Title:         ie.Title,
		Description:   ie.Description,
		Price:         ie.Price,
		Duration:      ie.Duration,
	}

	newContract, err := cont.svc.Add(c.Request().Context(), uc.Subject.ID, &dto)
	if err != nil {
		return fmt.Errorf("unable to save contract: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceContract, newContract.ID))
	return c.JSON(http.StatusCreated, newContract)
}

// @Summary     Get contract
// @Description Returns contract with specified id. This operation is allowed only for performer or customer.
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id} [get]
func (cont *Contract) get(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	id := c.Param("id")
	o, err := cont.svc.GetByIDForPerson(ctx, id, uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     List contracts
// @Description Returns list of contracts where the current user is performer or customer
// @Tags        contract
// @Accept      json
// @Produce     json
// @Success     200 {array}  model.ContractDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts [get]
func (cont *Contract) list(c echo.Context) error {
	ctx := c.Request().Context()

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	oo, err := cont.svc.ListByPersonID(ctx, uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, oo)
}

// @Summary     Accept contract
// @Description Performer is accepting contract
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/accept [post]
func (cont *Contract) accept(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Accept(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

type deployContractParams struct {
	ContractAddress string `json:"contract_address" validate:"required"`
}

// @Summary     Deploy contract
// @Description Customer has deployed contract to the blockchain
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/deploy [post]
func (cont *Contract) deploy(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(deployContractParams)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if err = validateStruct(ie); err != nil {
		return err
	}

	dto := model.DeployContractDTO{
		ContractAddress: ie.ContractAddress,
	}

	o, err := cont.svc.Deploy(c.Request().Context(), c.Param("id"), uc.Subject.ID, &dto)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Sign contract
// @Description Performer is signing contract
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/sign [post]
func (cont *Contract) sign(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Sign(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Fund contract
// @Description Customer is funded contract in the blockchain
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/fund [post]
func (cont *Contract) fund(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Fund(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Approve working results and allow to withdraw money from Smart Contract
// @Description Customer is approving performer's working results
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422 {object} model.BackendError "insufficient funds on the contract in the blockchain"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/approve [post]
func (cont *Contract) approve(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Approve(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Complete contract
// @Description Performer is completing contract
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.ContractDTO
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/complete [post]
func (cont *Contract) complete(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Complete(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

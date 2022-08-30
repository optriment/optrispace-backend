package controller

import (
	"fmt"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
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
	e.POST(resourceContract+"/:id/send", cont.send)
	e.POST(resourceContract+"/:id/approve", cont.approve)
	e.POST(resourceContract+"/:id/complete", cont.complete)
	log.Debug().Str("controller", resourceContract).Msg("Registered")
}

type contractDescription struct {
	PerformerID     string          `json:"performer_id,omitempty"`
	ApplicationID   string          `json:"application_id,omitempty"`
	Title           string          `json:"title,omitempty"`
	Description     string          `json:"description,omitempty"`
	Price           decimal.Decimal `json:"price,omitempty"`
	Duration        int32           `json:"duration,omitempty"`
	CustomerAddress string          `json:"customer_address,omitempty"`
}

// @Summary     Create a new contract
// @Description Creates a new contract based on existent application.
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       job body     controller.jobDescription true "New contract description"
// @Success     200    {object} model.Contract
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     409 {object} model.BackendError "duplication"
// @Failure     422 {object} model.BackendError "validation failed"
// @Failure     500    {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts [post]
func (cont *Contract) add(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(contractDescription)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.ApplicationID == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("application_id"),
		}
	}

	if ie.PerformerID == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("performer_id"),
		}
	}

	if ie.Title == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("title"),
		}
	}

	if ie.Description == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("description"),
		}
	}

	if decimal.Zero.Equal(ie.Price) {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("price"),
		}
	}

	if ie.Price.IsNegative() {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorMustBePositive("price"),
		}
	}

	if ie.CustomerAddress == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("customer_address"),
		}
	}

	newContract := &model.Contract{
		Customer:        &model.Person{ID: uc.Subject.ID},
		Performer:       &model.Person{ID: ie.PerformerID},
		Application:     &model.Application{ID: ie.ApplicationID},
		Title:           ie.Title,
		Description:     ie.Description,
		Price:           ie.Price,
		Duration:        ie.Duration,
		CreatedBy:       uc.Subject.ID,
		CustomerAddress: ie.CustomerAddress,
	}

	newContract, err = cont.svc.Add(c.Request().Context(), newContract)
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
// @Param       id     path     string                      true "Contract ID"
// @Success     200    {object} model.Contract
// @Failure     401    {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     500    {object} echo.HTTPError{message=string}
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
// @Success     200 {array}  model.Contract
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

type acceptParameters struct {
	PerformerAddress string `json:"performer_address,omitempty"`
}

// @Summary     Accept contract
// @Description Performer is accepting contract
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       params body     controller.acceptParameters true "Parameters"
// @Param       id     path     string                      true "Contract ID"
// @Success     200 {object} model.Contract
// @Failure     400    {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404    {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422    {object} model.BackendError "validation failed"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/accept [post]
func (cont *Contract) accept(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(acceptParameters)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.PerformerAddress == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("performer_address"),
		}
	}

	o, err := cont.svc.Accept(c.Request().Context(), c.Param("id"), uc.Subject.ID, ie.PerformerAddress)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

type deployParameters struct {
	ContractAddress string `json:"contract_address,omitempty"`
}

// @Summary     Deploy contract
// @Description Customer is deploying contract in the blockchain
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       params body     controller.deployParameters true "Parameters"
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.Contract
// @Failure     400    {object} model.BackendError "invalid format"
// @Failure     401    {object} model.BackendError "user not authorized"
// @Failure     403    {object} model.BackendError "insufficient rights"
// @Failure     404    {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     422    {object} model.BackendError "validation failed or insufficient funds on the contract in the blockchain"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/deploy [post]
func (cont *Contract) deploy(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(deployParameters)

	if e := c.Bind(ie); e != nil {
		return e
	}

	if ie.ContractAddress == "" {
		return &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("contract_address"),
		}
	}

	o, err := cont.svc.Deploy(c.Request().Context(), c.Param("id"), uc.Subject.ID, ie.ContractAddress)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Send working results to customer
// @Description Performer is sending working results to customer
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.Contract
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     404 {object} model.BackendError "contract not found or user not authorized to view contract"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /contracts/{id}/send [post]
func (cont *Contract) send(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	o, err := cont.svc.Send(c.Request().Context(), c.Param("id"), uc.Subject.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, o)
}

// @Summary     Approve working results
// @Description Customer is approving performer's working results
// @Tags        contract
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Contract ID"
// @Success     200 {object} model.Contract
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
// @Success     200 {object} model.Contract
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     403    {object} model.BackendError "insufficient rights"
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

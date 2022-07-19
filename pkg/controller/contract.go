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

func (cont *Contract) add(c echo.Context) error {
	type addingContract struct {
		PerformerID     string          `json:"performer_id,omitempty"`
		ApplicationID   string          `json:"application_id,omitempty"`
		Title           string          `json:"title,omitempty"`
		Description     string          `json:"description,omitempty"`
		Price           decimal.Decimal `json:"price,omitempty"`
		Duration        int32           `json:"duration,omitempty"`
		CustomerAddress string          `json:"customer_address,omitempty"`
	}

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(addingContract)

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

func (cont *Contract) accept(c echo.Context) error {
	type inputParameters struct {
		PerformerAddress string `json:"performer_address,omitempty"`
	}

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(inputParameters)

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

func (cont *Contract) deploy(c echo.Context) error {
	type inputParameters struct {
		ContractAddress string `json:"contract_address,omitempty"`
	}

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(inputParameters)

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

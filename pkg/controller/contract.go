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
		svc service.Contract
		sm  service.Security
	}
)

// NewContract create new service
func NewContract(svc service.Contract, sm service.Security) Registerer {
	return &Contract{
		svc: svc,
		sm:  sm,
	}
}

// Register implements Registerer interface
func (cont *Contract) Register(e *echo.Echo) {
	e.POST(resourceContract, cont.add)
	log.Debug().Str("controller", resourceContract).Msg("Registered")
}

func (cont *Contract) add(c echo.Context) error {
	type addingContract struct {
		PerformerID   string          `json:"performer_id,omitempty"`
		ApplicationID string          `json:"application_id,omitempty"`
		Title         string          `json:"title,omitempty"`
		Description   string          `json:"description,omitempty"`
		Price         decimal.Decimal `json:"price,omitempty"`
		Duration      int32           `json:"duration,omitempty"`
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
		return fmt.Errorf("application_id required: %w", model.ErrValueIsRequired)
	}

	if ie.PerformerID == "" {
		return fmt.Errorf("performer_id required: %w", model.ErrValueIsRequired)
	}

	if ie.Title == "" {
		return fmt.Errorf("title required: %w", model.ErrValueIsRequired)
	}

	if ie.Description == "" {
		return fmt.Errorf("description required: %w", model.ErrValueIsRequired)
	}

	if decimal.Zero.Equal(ie.Price) {
		return fmt.Errorf("price required: %w", model.ErrValueIsRequired)
	}

	if ie.Price.IsNegative() {
		return fmt.Errorf("price must be positive: %w", model.ErrInvalidValue)
	}

	newContract := &model.Contract{
		Customer:    &model.Person{ID: uc.Subject.ID},
		Performer:   &model.Person{ID: ie.PerformerID},
		Application: &model.Application{ID: ie.ApplicationID},
		Title:       ie.Title,
		Description: ie.Description,
		Price:       ie.Price,
		Duration:    ie.Duration,
		CreatedBy:   uc.Subject.ID,
	}

	newContract, err = cont.svc.Add(c.Request().Context(), newContract)
	if err != nil {
		return fmt.Errorf("unable to save contract: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourceContract, newContract.ID))
	return c.JSON(http.StatusCreated, newContract)
}

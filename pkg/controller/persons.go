package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/model"
	"optrispace.com/work/pkg/service"
)

type (
	// Person controller
	Person struct {
		sm  service.Security
		svc service.Person
	}
)

// NewPerson create new service
func NewPerson(sm service.Security, svc service.Person) Registerer {
	return &Person{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Person) Register(e *echo.Echo) {
	e.POST(resourcePerson, cont.add)
	e.GET(resourcePerson, cont.list)
	e.GET(resourcePerson+"/:id", cont.get)
	e.PUT(resourcePerson+"/:id", cont.update)
	e.PUT(resourcePerson+"/:id/resources", cont.setResources)
	log.Debug().Str("controller", resourcePerson).Msg("Registered")
}

// add is NOT signup method. For signup please see /signup endpoint!
// Warning! We should NOT add this method to the Swagger specification
func (cont *Person) add(c echo.Context) error {
	input := new(model.Person)

	if err := c.Bind(input); err != nil {
		return err
	}

	if uc, e := cont.sm.FromEchoContext(c); e != nil {
		return e
	} else if !uc.Subject.IsAdmin {
		return model.ErrInsufficientRights
	}

	input.ID = ""
	o, err := cont.svc.Add(c.Request().Context(), input)
	if err != nil {
		return fmt.Errorf("unable to save job: %w", err)
	}

	c.Response().Header().Set(echo.HeaderLocation, path.Join("/", resourcePerson, o.ID))
	return c.JSON(http.StatusCreated, o)
}

// @Summary     Get person description by id
// @Description Returns person description by id
// @Tags        person
// @Accept      json
// @Produce     json
// @Param       id  path     string true "Person ID"
// @Success     200 {object} model.Person
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "person not found"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /persons/{id} [get]
func (cont *Person) get(c echo.Context) error {
	id := c.Param("id")

	if uc, e := cont.sm.FromEchoContext(c); e != nil {
		return e
	} else if !uc.Subject.IsAdmin {
		return model.ErrInsufficientRights
	}

	o, err := cont.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, o)
}

// Warning! We should NOT add this method to the Swagger specification
func (cont *Person) list(c echo.Context) error {
	if uc, e := cont.sm.FromEchoContext(c); e != nil {
		return e
	} else if !uc.Subject.IsAdmin {
		return model.ErrInsufficientRights
	}

	oo, err := cont.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, oo)
}

type updatePerson struct {
	EthereumAddress string `json:"ethereum_address,omitempty"`
	DisplayName     string `json:"display_name,omitempty"`
	Email           string `json:"email,omitempty"`
}

var _ = updatePerson{}

// @Summary     Update existent person
// @Description Updates existent person. User must be authenticated as this person.
// @Tags        person
// @Accept      json
// @Produce     json
// @Param       job body     controller.updatePerson true "Update person details"
// @Param       id  path     string                  true "Person ID"
// @Success     200 {object} model.Person
// @Failure     400 {object} model.BackendError "invalid format"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "person not found"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /persons/{id} [put]
func (cont *Person) update(c echo.Context) error {
	ie := make(map[string]any)

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	if e := c.Bind(&ie); e != nil {
		return e
	}

	oo, err := cont.svc.Patch(c.Request().Context(), id, uc.Subject.ID, ie)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, oo)
}

// @Summary     Set resources for person
// @Description Updates resources for existent person. User must be authenticated as this person.
// @Tags        person
// @Accept      json
// @Produce     json
// @Param       job body string true "Person resources in JSON"
// @Param       id  path string true "Person ID"
// @Success     200
// @Failure     400 {object} model.BackendError "invalid format (body should be formatted as JSON)"
// @Failure     401 {object} model.BackendError "user not authorized"
// @Failure     404 {object} model.BackendError "person not found"
// @Failure     403 {object} model.BackendError "insufficient rights"
// @Failure     500 {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /persons/{id}/resources [put]
func (cont *Person) setResources(c echo.Context) error {
	ie := make(map[string]any)

	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}
	id := c.Param("id")

	bb, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return fmt.Errorf("unable to read request body: %w", model.ErrInvalidFormat)
	}
	if len(bb) == 0 {
		bb = []byte("{}")
	}

	// just for validation purposes
	if e := json.Unmarshal(bb, &ie); e != nil {
		return &model.BackendError{
			Cause:    model.ErrInvalidFormat,
			Message:  "body is not properly formatted json",
			TechInfo: e.Error(),
		}
	}

	err = cont.svc.SetResources(c.Request().Context(), id, uc.Subject.ID, bb)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, json.RawMessage("{}"))
}

package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"optrispace.com/work/pkg/service"
)

type (
	// Chat controller for messaging
	Chat struct {
		sm  service.Security
		svc service.Chat
	}
)

// NewChat create new service
func NewChat(sm service.Security, svc service.Chat) Registerer {
	return &Chat{
		sm:  sm,
		svc: svc,
	}
}

// Register implements Registerer interface
func (cont *Chat) Register(e *echo.Echo) {
	e.POST("/chats/:chat_id/messages", cont.addMessage)
	e.GET("/chats/:chat_id", cont.getChat)
	log.Debug().Str("controller", "chats").Msg("Registered")
}

type newMessage struct {
	Text string `json:"text,omitempty"`
}

// @Summary     Post a new message to the chat
// @Description A chat participant sending message to the chat
// @Tags        chat
// @Accept      json
// @Produce     json
// @Param       message body     controller.newMessage true "New message"
// @Param       chat_id path     string             true "chat id"
// @Success     201     {string} model.Message
// @Failure     401     {object} model.BackendError "user is not authorized"
// @Failure     403     {object} model.BackendError "user is not conversation participant"
// @Failure     404     {object} model.BackendError "chat does not exist"
// @Failure     422     {object} model.BackendError "message text exceeds maximum text length"
// @Failure     500     {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /chats/{chat_id}/messages [post]
func (cont *Chat) addMessage(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	ie := new(newMessage)

	if e := c.Bind(ie); e != nil {
		return e
	}

	m, err := cont.svc.AddMessage(c.Request().Context(), c.Param("chat_id"), uc.Subject.ID, ie.Text)
	if err == nil {
		return c.JSON(http.StatusCreated, m)
	}
	return err
}

// @Summary     Returns a fully chat description
// @Description A chat participant requesting chat description with all messages
// @Tags        chat
// @Accept      json
// @Produce     json
// @Param       chat_id path     string                true "chat id"
// @Success     200     {string} model.Chat         "chat will be returned with all messages"
// @Failure     401     {object} model.BackendError "user is not authorized"
// @Failure     403     {object} model.BackendError "user is not conversation participant"
// @Failure     404     {object} model.BackendError "chat does not exist"
// @Failure     500     {object} echo.HTTPError{message=string}
// @Security    BearerToken
// @Router      /chats/{chat_id} [get]
func (cont *Chat) getChat(c echo.Context) error {
	uc, err := cont.sm.FromEchoContext(c)
	if err != nil {
		return err
	}

	m, err := cont.svc.Get(c.Request().Context(), c.Param("chat_id"), uc.Subject.ID)
	if err == nil {
		return c.JSON(http.StatusOK, m)
	}
	return err
}

package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// ChatSvc is a application service
	ChatSvc struct {
		db *sql.DB
	}
)

const (
	topicApplicationPrefix = "urn:application:"
	topicContractPrefix    = "urn:contract:"

	messageTextMaxLen = 4096
)

var (
	chatTopicApplication = func(id string) string { return topicApplicationPrefix + id }
	chatTopicContract    = func(id string) string { return topicContractPrefix + id } //nolint: deadcode,unused,varcheck
)

// NewChat creates service
func NewChat(db *sql.DB) *ChatSvc {
	return &ChatSvc{db: db}
}

func chatFromDB(src pgdao.Chat, messages ...pgdao.Message) *model.Chat {
	dst := &model.Chat{
		ID:        src.ID,
		CreatedAt: src.CreatedAt,
		Topic:     src.Topic,
		Messages:  []model.Message{},
	}

	for _, m := range messages {
		dst.Messages = append(dst.Messages, model.Message{
			ID:        m.ID,
			ChatID:    src.ID,
			CreatedAt: m.CreatedAt,
			CreatedBy: m.CreatedBy,
			Text:      m.Text,
		})
	}
	return dst
}

// starts new chat and immediately add the first message to it
func newChat(ctx context.Context, queries *pgdao.Queries, topic, text, createdBy string, participants ...string) (*model.Chat, error) {
	newChat, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
		ID:    pgdao.NewID(),
		Topic: topic,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create new chat: %w", err)
	}

	_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
		ChatID:   newChat.ID,
		PersonID: createdBy,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to add a person %s for a new chat: %w", createdBy, err)
	}

	for _, p := range participants {
		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   newChat.ID,
			PersonID: p,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to add a person %s for a new chat: %w", p, err)
		}
	}

	// copy application comment to the first message
	msg, err := queries.MessageAdd(ctx, pgdao.MessageAddParams{
		ID:        pgdao.NewID(),
		ChatID:    newChat.ID,
		CreatedBy: createdBy,
		Text:      text,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create new message: %w", err)
	}

	return chatFromDB(newChat, msg), nil
}

// AddMessage implements service.Chat
func (s *ChatSvc) AddMessage(ctx context.Context, chatID, participantID, text string) (*model.Message, error) {
	if strings.TrimSpace(text) == "" {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorRequired("text"),
		}
	}

	if utf8.RuneCountInString(text) > messageTextMaxLen {
		return nil, &model.BackendError{
			Cause:   model.ErrValidationFailed,
			Message: model.ValidationErrorTooLong("text"),
		}
	}

	var result *model.Message
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		if _, e := queries.ChatParticipantGet(ctx, pgdao.ChatParticipantGetParams{
			ChatID:   chatID,
			PersonID: participantID,
		}); e != nil {
			if errors.Is(e, sql.ErrNoRows) {
				e = model.ErrInsufficientRights
			}
			return e
		}

		r, err := queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chatID,
			CreatedBy: participantID,
			Text:      strings.TrimSpace(text),
		})

		result = &model.Message{
			ID:        r.ID,
			ChatID:    r.ChatID,
			CreatedAt: r.CreatedAt,
			CreatedBy: r.CreatedBy,
			Text:      r.Text,
		}
		return err
	})
}

// Get implements service.Chat
func (s *ChatSvc) Get(ctx context.Context, chatID, participantID string) (*model.Chat, error) {
	var result *model.Chat
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		ch, err := queries.ChatGet(ctx, chatID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.ErrEntityNotFound
		}

		if err != nil {
			return err
		}

		if _, e := queries.ChatParticipantGet(ctx, pgdao.ChatParticipantGetParams{
			ChatID:   chatID,
			PersonID: participantID,
		}); e != nil {
			if errors.Is(e, sql.ErrNoRows) {
				e = model.ErrEntityNotFound
			}
			return e
		}

		mm, err := queries.MessagesListByChat(ctx, chatID)
		if err != nil {
			return err
		}

		result = chatFromDB(ch)
		for _, m := range mm {
			result.Messages = append(result.Messages, model.Message{
				ID:         m.ID,
				ChatID:     chatID,
				CreatedAt:  m.CreatedAt,
				CreatedBy:  m.CreatedBy,
				AuthorName: m.DisplayName,
				Text:       m.Text,
			})
		}

		return nil
	})
}

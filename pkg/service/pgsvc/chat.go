package pgsvc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

type (
	// ChatSvc is a application service
	ChatSvc struct {
		db *sql.DB
	}
)

func topicParts(topic string) (kind, id string) {
	switch {
	case strings.HasPrefix(topic, "urn:"+kindApplication+":"):
		return kindApplication, topic[len(kindApplication)+5:]
	case strings.HasPrefix(topic, "urn:"+kindContract+":"):
		return kindContract, topic[len(kindContract)+5:]
	}
	return "", ""
}

const (
	kindApplication = "application"
	kindContract    = "contract"

	messageTextMaxLen = 4096
)

var (
	newChatTopicApplication = func(id string) string { return "urn:" + kindApplication + ":" + id }
	newChatTopicContract    = func(id string) string { return "urn:" + kindContract + ":" + id } //nolint: deadcode,unused,varcheck
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
			return nil, fmt.Errorf("unable to add a participant %s for a new chat: %w", p, err)
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

// ListByParticipant implements service.Chat
func (s *ChatSvc) ListByParticipant(ctx context.Context, participantID string) ([]*model.ChatDTO, error) {
	var result []*model.ChatDTO
	return result, doWithQueries(ctx, s.db, defaultRwTxOpts, func(queries *pgdao.Queries) error {
		cc, err := queries.ChatsListByParticipant(ctx, participantID)
		if err != nil {
			return err
		}

		for _, c := range cc {

			var (
				kind, id      = topicParts(c.Topic)
				jobID         = ""
				applicationID = ""
				contractID    = ""
				title         = ""
			)

			switch kind {
			case kindApplication:
				details, err := queries.ChatGetDetailsByApplicationID(ctx, id)
				if err != nil {
					clog.Ctx(ctx).Warn().Err(err).Str("chat-topic", c.Topic).Msg("Failed to get information about chat.")
				} else {
					jobID = details.JobID
					applicationID = details.ApplicationID
					contractID = details.ContractID.String
					title = details.JobTitle
				}
			}

			found := false
			for i := range result {
				if result[i].ID == c.ID {
					result[i].Participants = append(result[i].Participants, &model.ParticipantDTO{
						ID:              c.PersonID,
						DisplayName:     c.PersonDisplayName,
						EthereumAddress: c.PersonEthereumAddress,
					})
					found = true
					break
				}
			}
			if !found {
				result = append(result, &model.ChatDTO{
					ID:            c.ID,
					Topic:         c.Topic,
					Kind:          kind,
					Title:         title,
					JobID:         jobID,
					ApplicationID: applicationID,
					ContractID:    contractID,
					Participants: []*model.ParticipantDTO{
						{
							ID:              c.PersonID,
							DisplayName:     c.PersonDisplayName,
							EthereumAddress: c.PersonEthereumAddress,
						},
					},
				})
			}
		}

		return nil
	})
}

package intest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"optrispace.com/work/pkg/clog"
	"optrispace.com/work/pkg/db/pgdao"
	"optrispace.com/work/pkg/model"
)

var (
	chatsResourceName = "chats"
	chatsURL          = appURL + "/" + chatsResourceName
)

func TestApplicationChat(t *testing.T) {
	require.NoError(t, pgdao.PurgeDB(ctx, db))

	var (
		customer   = addPersonWithEthereumAddress(t, "customer", "0x8Ca2702c5bcc50D79d9a059D58607028aa36Aa6c")
		customer2  = addPersonWithEthereumAddress(t, "customer2", "0x8Ca2702c5bcc50D79d9a059D58607028aa36Aa77")
		performer  = addPersonWithEthereumAddress(t, "performer", "0x8Ca2702c5bcc50D79d9a059D58607028aa36Aa6c")
		performer2 = addPersonWithEthereumAddress(t, "performer2", "0x8Ca2702c5bcc50D79d9a059D58607028aa36Aa78")
		stranger   = addPerson(t, "stranger")

		job1 = addJob(t, "A job1", "Some beautiful 1", customer.ID, "", "")
		job2 = addJob(t, "A job2", "Some beautiful 2", customer.ID, "", "")
		job3 = addJob(t, "A job3", "Some beautiful 3", customer.ID, "", "")
		job4 = addJob(t, "A job4", "Some beautiful 4", customer.ID, "", "")
		job5 = addJob(t, "A job5", "Some beautiful 5", customer2.ID, "", "")

		existentApplication1 = addApplication(t, job1.ID, "I need this job 1", "3.0", performer.ID)
		existentApplication2 = addApplication(t, job2.ID, "I need this job 2", "4.5", performer.ID)
	)

	t.Run("stranger get chat for an existent application", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL+"/applications/"+existentApplication1.ID+"/chat", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+stranger.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "insufficient rights", e.Message)
		}
	})

	t.Run("performer get chat for an existent application", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL+"/applications/"+existentApplication1.ID+"/chat", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Chat)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			pg := pgdao.ChatParticipantGetParams{ChatID: e.ID, PersonID: performer.ID}
			_, err := queries.ChatParticipantGet(ctx, pg)
			require.NoError(t, err)

			pg.PersonID = customer.ID
			_, err = queries.ChatParticipantGet(ctx, pg)
			require.NoError(t, err)

			mm, err := queries.MessagesListByChat(ctx, e.ID)
			require.NoError(t, err)

			if assert.Len(t, mm, 1) {
				assert.Equal(t, "I need this job 1", mm[0].Text)
			}

		}
	})

	t.Run("customer get chat for an existent application", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, appURL+"/applications/"+existentApplication2.ID+"/chat", nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Chat)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			mm, err := queries.MessagesListByChat(ctx, e.ID)
			require.NoError(t, err)

			if assert.Len(t, mm, 1) {
				assert.Equal(t, "I need this job 2", mm[0].Text)
			}

		}
	})

	t.Run("create chat while user is applying for a job", func(t *testing.T) {
		body := `{
			"comment":"My awesome comment",
			"price": "44.77895"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/jobs/"+job3.ID+"/applications", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+performer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			chat, err := queries.ChatGetByTopic(ctx, "urn:application:"+e.ID)
			require.NoError(t, err)

			mm, err := queries.MessagesListByChat(ctx, chat.ID)
			require.NoError(t, err)

			if assert.Len(t, mm, 1) {
				assert.NotEmpty(t, mm[0].ID)
				assert.Equal(t, "My awesome comment", mm[0].Text)
				assert.NotEmpty(t, mm[0].CreatedAt)
				assert.Equal(t, performer.ID, mm[0].CreatedBy)
			}
		}
	})

	t.Run("send 2 message for application", func(t *testing.T) {
		appl := doRequest[model.Application](t, http.MethodPost, appURL+"/jobs/"+job4.ID+"/applications",
			`{"comment":"Me, please!","price": "22.11"}`, performer.AccessToken.String)

		chat := doRequest[model.Chat](t, http.MethodGet, appURL+"/applications/"+appl.ID+"/chat",
			``, performer.AccessToken.String)

		assert.Equal(t, "urn:application:"+appl.ID, chat.Topic)

		_ = doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat.ID+"/messages", `{"text":"customer is questioning"}`, customer.AccessToken.String)
		_ = doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat.ID+"/messages", `{"text":"performer is replying"}`, performer.AccessToken.String)

		updatedChat := doRequest[model.Chat](t, http.MethodGet, appURL+"/chats/"+chat.ID, ``, customer.AccessToken.String)

		if assert.Len(t, updatedChat.Messages, 3) {

			mm := updatedChat.Messages
			assert.Equal(t, "Me, please!", mm[0].Text)
			assert.Equal(t, "customer is questioning", mm[1].Text)
			assert.Equal(t, "performer is replying", mm[2].Text)

			assert.Equal(t, performer.ID, mm[0].CreatedBy)
			assert.Equal(t, customer.ID, mm[1].CreatedBy)
			assert.Equal(t, performer.ID, mm[2].CreatedBy)

			assert.NotEmpty(t, mm[0].CreatedAt)
			assert.NotEmpty(t, mm[1].CreatedAt)
			assert.NotEmpty(t, mm[2].CreatedAt)

			assert.NotEmpty(t, mm[0].ID)
			assert.NotEmpty(t, mm[1].ID)
			assert.NotEmpty(t, mm[2].ID)

			assert.Equal(t, updatedChat.ID, mm[0].ChatID)
			assert.Equal(t, updatedChat.ID, mm[1].ChatID)
			assert.Equal(t, updatedChat.ID, mm[2].ChatID)
		}
	})

	t.Run("performer 2 gets chat list", func(t *testing.T) {
		appl4 := doRequest[model.Application](t, http.MethodPost, appURL+"/jobs/"+job4.ID+"/applications",
			`{"comment":"I want to develop this job4. Yes, I will.","price": "44.33"}`, performer2.AccessToken.String)

		appl5 := doRequest[model.Application](t, http.MethodPost, appURL+"/jobs/"+job5.ID+"/applications",
			`{"comment":"I want to develop this job5.","price": "55.11"}`, performer2.AccessToken.String)

		chat4 := doRequest[model.Chat](t, http.MethodGet, appURL+"/applications/"+appl4.ID+"/chat", "", customer.AccessToken.String)
		chat5 := doRequest[model.Chat](t, http.MethodGet, appURL+"/applications/"+appl5.ID+"/chat", "", customer2.AccessToken.String)

		_ = doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat4.ID+"/messages", `{"text":"customer is questioning"}`, customer.AccessToken.String)
		message2 := doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat5.ID+"/messages", `{"text":"customer2 is questioning"}`, customer2.AccessToken.String)
		message3 := doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat4.ID+"/messages", `{"text":"performer2 is replying"}`, performer2.AccessToken.String)

		chats := doRequest[[]*model.ChatDTO](t, http.MethodGet, appURL+"/chats", "", performer2.AccessToken.String)

		require.Len(t, chats, 2)

		customerDTO := &model.ParticipantDTO{
			ID:              customer.ID,
			DisplayName:     customer.DisplayName,
			EthereumAddress: customer.EthereumAddress,
		}

		customer2DTO := &model.ParticipantDTO{
			ID:              customer2.ID,
			DisplayName:     customer2.DisplayName,
			EthereumAddress: customer2.EthereumAddress,
		}

		performer2DTO := &model.ParticipantDTO{
			ID:              performer2.ID,
			DisplayName:     performer2.DisplayName,
			EthereumAddress: performer2.EthereumAddress,
		}

		c := chats[0]
		assert.NotEmpty(t, c.ID)
		assert.Equal(t, "urn:application:"+appl4.ID, c.Topic)
		assert.Equal(t, "application", c.Kind)
		assert.Equal(t, "A job4", c.Title)
		assert.Equal(t, job4.ID, c.JobID)
		assert.Equal(t, appl4.ID, c.ApplicationID)
		assert.Empty(t, c.ContractID)
		assert.Equal(t, c.LastMessageAt, message3.CreatedAt)
		if assert.Len(t, c.Participants, 2) {
			assert.Contains(t, c.Participants, performer2DTO)
			assert.Contains(t, c.Participants, customerDTO)
		}

		c = chats[1]
		assert.NotEmpty(t, c.ID)
		assert.Equal(t, "urn:application:"+appl5.ID, c.Topic)
		assert.Equal(t, "application", c.Kind)
		assert.Equal(t, "A job5", c.Title)
		assert.Equal(t, job5.ID, c.JobID)
		assert.Equal(t, appl5.ID, c.ApplicationID)
		assert.Empty(t, c.ContractID)
		assert.Equal(t, c.LastMessageAt, message2.CreatedAt)
		if assert.Len(t, c.Participants, 2) {
			assert.Contains(t, c.Participants, performer2DTO)
			assert.Contains(t, c.Participants, customer2DTO)
		}
	})
}

func TestCreateChat(t *testing.T) {
	t.Run("creates a new chat while applying for a job", func(t *testing.T) {
		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:        pgdao.NewID(),
			CreatedBy: customer.ID,
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
			EthereumAddress: "0x1234567890APPLICANT",
		})
		require.NoError(t, err)

		body := `{
			"comment":" My awesome comment\n ",
			"price": "44.77895"
		}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/jobs/"+job.ID+"/applications", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Application)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			chat, err := queries.ChatGetByTopic(ctx, "urn:application:"+e.ID)
			require.NoError(t, err)

			messages, err := queries.MessagesListByChat(ctx, chat.ID)
			require.NoError(t, err)

			if assert.Len(t, messages, 1) {
				assert.NotEmpty(t, messages[0].ID)
				assert.Equal(t, "My awesome comment", messages[0].Text)
				assert.NotEmpty(t, messages[0].CreatedAt)
				assert.Equal(t, applicant.ID, messages[0].CreatedBy)
			}
		}
	})
}

func TestSendMessageToChat(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		body := `{}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/chats/12345/messages", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("with validation errors", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		chatMessagesURL := appURL + "/" + chatsResourceName + "/12345/messages"

		t.Run("returns error if body is not a valid JSON", func(t *testing.T) {
			body := `{z}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatMessagesURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "invalid JSON", e["message"])
			}
		})

		t.Run("returns error if text is missing", func(t *testing.T) {
			body := `{}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatMessagesURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "text is required", e["message"])
			}
		})

		t.Run("returns error if text is an empty string", func(t *testing.T) {
			body := `{"text":" "}`

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatMessagesURL, bytes.NewReader([]byte(body)))
			require.NoError(t, err)
			req.Header.Set(clog.HeaderXHint, t.Name())
			req.Header.Set(echo.HeaderContentType, "application/json")
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode, "Invalid result status code '%s'", res.Status) {
				e := map[string]any{}
				require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
				assert.Equal(t, "text is required", e["message"])
			}
		})
	})

	t.Run("returns error if chat does not exist", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{"text":"message"}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/chats/12345/messages", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "entity not found", e["message"])
		}
	})

	t.Run("returns error if person is not a participant of this chat", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: customer.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: applicant.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		body := `{"text":"message"}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/chats/" + chat.ID + "/messages", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusForbidden, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "insufficient rights", e["message"])
		}
	})

	t.Run("adds message to chat as an applicant", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: customer.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: applicant.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		body := `{"text":" Second Message "}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/chats/" + chat.ID + "/messages", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Message)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, chat.ID, e.ChatID)
				assert.Equal(t, applicant.ID, e.CreatedBy)
				assert.Equal(t, "Second Message", e.Text)
				assert.NotEmpty(t, e.CreatedAt)
			}
		}
	})

	t.Run("adds message to chat as a customer", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: customer.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: applicant.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		body := `{"text":" Reply "}`

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, appURL+"/chats/" + chat.ID + "/messages", bytes.NewBufferString(body))
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusCreated, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := new(model.Message)
			require.NoError(t, json.NewDecoder(res.Body).Decode(e))

			if assert.NotEmpty(t, e) {
				assert.NotEmpty(t, e.ID)
				assert.Equal(t, chat.ID, e.ChatID)
				assert.Equal(t, customer.ID, e.CreatedBy)
				assert.Equal(t, "Reply", e.Text)
				assert.NotEmpty(t, e.CreatedAt)
			}
		}
	})
}

func TestGetChats(t *testing.T) {
	t.Run("returns error for unauthorized request", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, chatsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := map[string]any{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.Equal(t, "Authorization required", e["message"])
		}
	})

	t.Run("returns an empty array if there are no chats for the person", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		person, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, chatsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+person.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ChatDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))
			assert.Empty(t, ee)
		}
	})

	t.Run("returns only related chats for the customer", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: customer.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: applicant.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		lastMessage, err := queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: customer.ID,
			Text:      "second message",
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer2.ID,
		})
		require.NoError(t, err)

		application2, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job2.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat2, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application2.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat2.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat2.ID,
			PersonID: customer2.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat2.ID,
			CreatedBy: applicant.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, chatsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ChatDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			assert.Equal(t, 1, len(ee))

			expectedChat := ee[0]

			assert.Equal(t, expectedChat.ID, chat.ID)
			assert.Equal(t, expectedChat.Topic, "urn:application:" + application.ID)
			assert.Equal(t, expectedChat.Kind, "application")
			assert.Equal(t, expectedChat.Title, "Title")
			assert.Equal(t, expectedChat.JobID, job.ID)
			assert.Equal(t, expectedChat.ApplicationID, application.ID)
			assert.Empty(t, expectedChat.ContractID)
			assert.Equal(t, expectedChat.LastMessageAt.UTC(), lastMessage.CreatedAt.UTC())

			customerDTO := &model.ParticipantDTO{
				ID:              customer.ID,
				DisplayName:     customer.DisplayName,
				EthereumAddress: customer.EthereumAddress,
			}

			applicantDTO := &model.ParticipantDTO{
				ID:              applicant.ID,
				DisplayName:     applicant.DisplayName,
				EthereumAddress: applicant.EthereumAddress,
			}

			if assert.Len(t, expectedChat.Participants, 2) {
				assert.Contains(t, expectedChat.Participants, customerDTO)
				assert.Contains(t, expectedChat.Participants, applicantDTO)
			}
		}
	})

	t.Run("returns only related chats for the applicant", func(t *testing.T) {
		require.NoError(t, pgdao.PurgeDB(ctx, db))

		customer, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title",
			Description: "Description",
			CreatedBy:   customer.ID,
		})
		require.NoError(t, err)

		applicant, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
			AccessToken: sql.NullString{
				String: pgdao.NewID(),
				Valid:  true,
			},
		})
		require.NoError(t, err)

		application, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat.ID,
			PersonID: customer.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: applicant.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		lastMessage, err := queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat.ID,
			CreatedBy: customer.ID,
			Text:      "second message",
		})
		require.NoError(t, err)

		customer2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		job2, err := pgdao.New(db).JobAdd(ctx, pgdao.JobAddParams{
			ID:          pgdao.NewID(),
			Title:       "Title 2",
			Description: "Description",
			CreatedBy:   customer2.ID,
		})
		require.NoError(t, err)

		application2, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job2.ID,
			ApplicantID: applicant.ID,
		})
		require.NoError(t, err)

		chat2, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application2.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat2.ID,
			PersonID: applicant.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat2.ID,
			PersonID: customer2.ID,
		})
		require.NoError(t, err)

		lastMessage2, err := queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat2.ID,
			CreatedBy: applicant.ID,
			Text:      "another message",
		})
		require.NoError(t, err)

		applicant2, err := pgdao.New(db).PersonAdd(ctx, pgdao.PersonAddParams{
			ID:    pgdao.NewID(),
			Login: pgdao.NewID(),
		})
		require.NoError(t, err)

		application3, err := pgdao.New(db).ApplicationAdd(ctx, pgdao.ApplicationAddParams{
			ID:          pgdao.NewID(),
			Comment:     "Comment 1",
			Price:       "42.35",
			JobID:       job2.ID,
			ApplicantID: applicant2.ID,
		})
		require.NoError(t, err)

		chat3, err := queries.ChatAdd(ctx, pgdao.ChatAddParams{
			ID:    pgdao.NewID(),
			Topic: "urn:application:" + application3.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat3.ID,
			PersonID: applicant2.ID,
		})
		require.NoError(t, err)

		_, err = queries.ChatParticipantAdd(ctx, pgdao.ChatParticipantAddParams{
			ChatID:   chat3.ID,
			PersonID: customer2.ID,
		})
		require.NoError(t, err)

		_, err = queries.MessageAdd(ctx, pgdao.MessageAddParams{
			ID:        pgdao.NewID(),
			ChatID:    chat3.ID,
			CreatedBy: applicant2.ID,
			Text:      "first message",
		})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, chatsURL, nil)
		require.NoError(t, err)
		req.Header.Set(clog.HeaderXHint, t.Name())
		req.Header.Set(echo.HeaderContentType, "application/json")
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+applicant.AccessToken.String)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if assert.Equal(t, http.StatusOK, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			ee := make([]*model.ChatDTO, 0)
			require.NoError(t, json.NewDecoder(res.Body).Decode(&ee))

			assert.Equal(t, 2, len(ee))

			firstChat := ee[0]

			assert.Equal(t, firstChat.ID, chat2.ID)
			assert.Equal(t, firstChat.Topic, "urn:application:" + application2.ID)
			assert.Equal(t, firstChat.Kind, "application")
			assert.Equal(t, firstChat.Title, "Title 2")
			assert.Equal(t, firstChat.JobID, job2.ID)
			assert.Equal(t, firstChat.ApplicationID, application2.ID)
			assert.Empty(t, firstChat.ContractID)
			assert.Equal(t, firstChat.LastMessageAt.UTC(), lastMessage2.CreatedAt.UTC())

			customer2DTO := &model.ParticipantDTO{
				ID:              customer2.ID,
				DisplayName:     customer2.DisplayName,
				EthereumAddress: customer2.EthereumAddress,
			}

			applicantDTO := &model.ParticipantDTO{
				ID:              applicant.ID,
				DisplayName:     applicant.DisplayName,
				EthereumAddress: applicant.EthereumAddress,
			}

			if assert.Len(t, firstChat.Participants, 2) {
				assert.Contains(t, firstChat.Participants, customer2DTO)
				assert.Contains(t, firstChat.Participants, applicantDTO)
			}

			secondChat := ee[1]

			assert.Equal(t, secondChat.ID, chat.ID)
			assert.Equal(t, secondChat.Topic, "urn:application:" + application.ID)
			assert.Equal(t, secondChat.Kind, "application")
			assert.Equal(t, secondChat.Title, "Title")
			assert.Equal(t, secondChat.JobID, job.ID)
			assert.Equal(t, secondChat.ApplicationID, application.ID)
			assert.Empty(t, secondChat.ContractID)
			assert.Equal(t, secondChat.LastMessageAt.UTC(), lastMessage.CreatedAt.UTC())

			customerDTO := &model.ParticipantDTO{
				ID:              customer.ID,
				DisplayName:     customer.DisplayName,
				EthereumAddress: customer.EthereumAddress,
			}

			if assert.Len(t, secondChat.Participants, 2) {
				assert.Contains(t, secondChat.Participants, customerDTO)
				assert.Contains(t, secondChat.Participants, applicantDTO)
			}
		}
	})
}

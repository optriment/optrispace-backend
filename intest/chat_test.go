package intest

import (
	"bytes"
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

func TestApplicationChat(t *testing.T) {
	pgdao.PurgeDB(ctx, db)

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

		if assert.Equal(t, http.StatusNotFound, res.StatusCode, "Invalid result status code '%s'", res.Status) {
			e := model.BackendError{}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&e))
			assert.EqualValues(t, "entity not found", e.Message)
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
		_ = doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat5.ID+"/messages", `{"text":"customer2 is questioning"}`, customer2.AccessToken.String)
		_ = doRequest[model.Message](t, http.MethodPost, appURL+"/chats/"+chat4.ID+"/messages", `{"text":"performer2 is replying"}`, performer2.AccessToken.String)

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
		if assert.Len(t, c.Participants, 2) {
			assert.Contains(t, c.Participants, performer2DTO)
			assert.Contains(t, c.Participants, customer2DTO)
		}
	})
}

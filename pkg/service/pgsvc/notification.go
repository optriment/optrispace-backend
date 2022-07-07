package pgsvc

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	// NotificationSvc is notification service
	NotificationSvc struct {
		token   string
		chatIDs []int64
		botAPI  *tgbotapi.BotAPI
	}
)

// NewNotification creates service
func NewNotification(tgToken string, chatIDs ...int64) *NotificationSvc {
	return &NotificationSvc{
		token:   tgToken,
		chatIDs: chatIDs,
	}
}

var markdownV2Replacer = strings.NewReplacer("\\", "\\\\", "`", "\\`")

type errs struct {
	errs []error
}

func (ee errs) hasErrs() bool {
	return len(ee.errs) > 0
}

func (ee errs) Error() string {
	sb := strings.Builder{}
	for _, e := range ee.errs {
		sb.WriteString(e.Error())
		sb.WriteString("; ")
	}
	return sb.String()
}

// Push implements Notification interface
// Sending is continue even if errors occur. Except error Telegram Bot API instance creating
func (s *NotificationSvc) Push(ctx context.Context, data string) error {
	if s.botAPI == nil {

		botAPI, e := tgbotapi.NewBotAPI(s.token)
		if e != nil {
			return fmt.Errorf("unable to create Telegram bot API: %w", e)
		}
		s.botAPI = botAPI
	}

	var ee errs
	for _, c := range s.chatIDs {
		m := tgbotapi.NewMessage(c, "```"+markdownV2Replacer.Replace(data)+"```")
		m.ParseMode = "MarkdownV2" // https://core.telegram.org/bots/api#markdownv2-style
		if _, err := s.botAPI.Send(m); err != nil {
			ee.errs = append(ee.errs, err)
		}
	}

	if ee.hasErrs() {
		return ee
	}
	return nil
}

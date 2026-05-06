package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const telegramHTTPTimeout = 60 * time.Second

// TelegramOutput sends text messages to one or more Telegram chats.
type TelegramOutput struct {
	bot     *tgbotapi.BotAPI
	roomIDs []int64
}

// NewTelegramOutput builds a TelegramOutput from an API token and comma-separated chat IDs.
func NewTelegramOutput(apiToken string, roomIDsCSV string) (*TelegramOutput, error) {
	ids, err := parseRoomIDs(roomIDsCSV)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Timeout: telegramHTTPTimeout}
	bot, err := tgbotapi.NewBotAPIWithClient(apiToken, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	return &TelegramOutput{bot: bot, roomIDs: ids}, nil
}

func parseRoomIDs(roomIDsCSV string) ([]int64, error) {
	parts := strings.Split(roomIDsCSV, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse room id %q: %w", p, err)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no telegram room ids configured")
	}
	return ids, nil
}

// SendMessage delivers text to every configured chat. ctx is checked before each send;
// HTTP calls use the client timeout configured in NewTelegramOutput.
func (t *TelegramOutput) SendMessage(ctx context.Context, text string) error {
	if t == nil || t.bot == nil {
		return fmt.Errorf("telegram: not initialized")
	}
	if len(t.roomIDs) == 0 {
		return fmt.Errorf("telegram: no chat ids configured")
	}
	for _, roomID := range t.roomIDs {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("telegram: %w", err)
		}
		msg := tgbotapi.NewMessage(roomID, text)
		if _, err := t.bot.Send(msg); err != nil {
			return fmt.Errorf("telegram: send to %d: %w", roomID, err)
		}
	}
	return nil
}

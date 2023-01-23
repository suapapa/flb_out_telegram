package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

var (
	tgBot   *tgbotapi.BotAPI
	tgRooms []int64
)

// initTgBot initialize telegram bot
//
// apiToken: telegram bot api token
// roomIDs: telegram room id (slice of string)
func initTgBot(apiToken string, roomIDs string) error {
	if tgBot != nil {
		return fmt.Errorf("tgBot already initialized")
	}

	var err error
	tgBot, err = tgbotapi.NewBotAPI(apiToken)
	if err != nil {
		return errors.Wrap(err, "fail to init telegram bot")
	}

	for _, roomID := range strings.Split(roomIDs, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(roomID), 10, 64)
		if err != nil {
			return errors.Wrap(err, "fail to init telegram bot")
		}
		tgRooms = append(tgRooms, id)
	}
	return nil
}

// sendMsgToTelegram send msg to telegram multiple rooms
func sendMsgToTelegram(msg string) error {
	if len(tgRooms) == 0 {
		return fmt.Errorf("no telegram room to send")
	}

	for _, roomID := range tgRooms {
		c := tgbotapi.NewMessage(roomID, msg)
		if _, err := tgBot.Send(c); err != nil {
			return errors.Wrap(err, "fail to send msg to telegram")
		}
	}

	return nil
}

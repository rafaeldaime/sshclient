package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type User struct {
	ChatID int64
	Host   string
	Port   int
	User   string
	Pass   string
	Bot    *tgbotapi.BotAPI
	State  State // What interface user is using
}

// int = User.ID
var users = map[int]*User{}

func NewUser(bot *tgbotapi.BotAPI, update tgbotapi.Update) *User {
	u := &User{Bot: bot, ChatID: GetChatID(update), Port: 22}
	users[GetUserID(update)] = u
	return u
}

func (u *User) Println(line string) (tgbotapi.Message, error) {
	m := tgbotapi.NewMessage(u.ChatID, line)
	m.DisableWebPagePreview = true
	return u.Bot.Send(m)
}

func (u *User) Send(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	return u.Bot.Send(msg)
}

func (u *User) HideKeyboard(text string) (tgbotapi.Message, error) {

	m := tgbotapi.NewMessage(u.ChatID, text)
	m.ReplyMarkup = tgbotapi.NewHideKeyboard(true)

	return u.Send(m)
}

func (u *User) IsReady() bool {
	if u.Host != "" && u.Port != 0 && u.User != "" && u.Pass != "" {
		return true
	}
	return false
}

package main

import (
	"io/ioutil"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"golang.org/x/crypto/ssh"
)

func GetUser(update tgbotapi.Update) *User {

	u, ok := users[GetUserID(update)]
	if !ok {
		return nil
	}
	return u
}

func GetUserID(update tgbotapi.Update) int {

	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

func GetChatID(update tgbotapi.Update) int64 {

	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func EditMessage(u *User, msg tgbotapi.Message, text string) (tgbotapi.Message, error) {

	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    u.ChatID,
			MessageID: msg.MessageID,
		},
		Text: text,
	}

	return u.Bot.Send(edit)
}

func Run(u *User, cmd string) error {

	sshConfig := &ssh.ClientConfig{
		User: u.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(u.Pass),
		},
	}

	client, err := ssh.Dial("tcp", u.Host+":"+strconv.Itoa(u.Port), sshConfig)
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	outPipe, err := session.StdoutPipe()
	if err != nil {
		u.Println("Error getting outPipe: " + err.Error())
	}

	errPipe, err := session.StderrPipe()
	if err != nil {
		u.Println("Error getting errPipe: " + err.Error())
	}

	err = session.Run(cmd)
	if err != nil {
		u.Println(err.Error())
	}

	stdOut, err := ioutil.ReadAll(outPipe)
	if err != nil {
		u.Println("Error reading stdOut: " + err.Error())
	}

	u.Println(string(stdOut))

	stdErr, err := ioutil.ReadAll(errPipe)
	if err != nil {
		u.Println("Error reading stdErr: " + err.Error())
	}

	u.Println(string(stdErr))

	return nil
}

package main

import (
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func Handle(u *User, update tgbotapi.Update) {
	switch u.State {
	case InitState:
		InitHandler(u, update)
	case ConfigState:
		ConfigHandler(u, update)
	case ReadyState:
		ReadyHandler(u, update)
	}

	// Insert instructions for next command
	PostProcess(u)
}

func PostProcess(u *User) {
	switch u.State {
	case InitState:
		InitHeader(u)
	case ConfigState:
		ConfigHeader(u)
	case ReadyState:
		ReadyHeader(u)
	}

}

func InitHeader(u *User) {
}

func InitHandler(u *User, update tgbotapi.Update) {

	switch update.Message.Text {
	case "/start", "Start":
		u.Println("OK, let's configure the SSH Client")
		u.Println("Keep calm, we don't store any data inserted by you")
		u.HideKeyboard("Type /back if you wanna go back")
		u.State = ConfigState

	case "/help", "Help":
		u.Println("Type /start to start the SSH Client")
		u.Println("Or /about to know more about this project")
		SendKeyboard(u, "Or you can just use the keyboard bellow")

	case "/about", "About":
		u.Println("This is the first SSH Client for Telegram to rapidly connect to your remote server with all messages encrypted by Telegram")
		u.Println("It's an open-source project found in github.com/rafadev7/sshclient")
		u.Println("We don't store any information you send through the very secure Telegram cryptography system")
		u.Println("If you got interested then access our github pages and get involved with the project")
		SendKeyboard(u, "Chose one of the options in the keyboard bellow")

	case "/rate", "Please give us five Stars!":
		SendRateInline(u)
		SendKeyboard(u, "Then you can choose any options below")

	default:
		SendKeyboard(u, "Welcome to the SSH Client for Telegram")
	}

}

func ConfigHeader(u *User) {
	if u.Host == "" {
		u.HideKeyboard("Insert your Host address:")
		return
	}

	if u.Port == 0 {
		u.HideKeyboard("Insert the Host Port:")
		return
	}

	if u.User == "" {
		u.HideKeyboard("Insert your User:")
		return
	}

	if u.Pass == "" {
		u.HideKeyboard("Insert your Password:")
		return
	}

	SendKeyboard(u, "Seems everything is configured")
}

func ConfigHandler(u *User, update tgbotapi.Update) {

	if update.Message.Text == "/back" {
		u.State = InitState
		SendKeyboard(u, "Welcome to the SSH Client for Telegram")
		return
	}

	if u.Host == "" {
		u.Host = update.Message.Text
		return
	}

	if u.Port == 0 {
		port, err := strconv.Atoi(update.Message.Text)
		if err != nil || port > 65535 {
			u.Println("Ops, your port should be a number and < 65535")
		}
		u.Port = port
		return
	}

	if u.User == "" {
		u.User = update.Message.Text
		return
	}

	if u.Pass == "" {
		u.Pass = update.Message.Text
		return
	}

	switch update.Message.Text {
	case "/host", "Host":
		u.Host = ""

	case "/port", "Port":
		u.Port = 0

	case "/user", "User":
		u.User = ""

	case "/pass", "Pass":
		u.Pass = ""

	case "/run", "Start sending commands!":
		u.HideKeyboard("Just type the commands you need to send...")
		u.HideKeyboard("Or type /back if you wanna go back")
		u.State = ReadyState

	}
}

func ReadyHeader(u *User) {
}

func ReadyHandler(u *User, update tgbotapi.Update) {

	if update.Message.Text == "/back" {
		u.State = ConfigState
		return
	}

	err := Run(u, update.Message.Text)
	if err != nil {
		u.Println("Fail:" + err.Error())
		u.Println("Something goes wrong connecting to the server")
		u.Println("You have to reconfigure the access data")
		u.State = ConfigState
		return
	}
}

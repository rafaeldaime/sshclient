package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

func SendKeyboard(u *User, text string) {

	message := tgbotapi.NewMessage(u.ChatID, text)

	switch u.State {
	case InitState:

		btnStart := tgbotapi.NewKeyboardButton("Start")
		btnAbout := tgbotapi.NewKeyboardButton("About")
		btnRow1 := tgbotapi.NewKeyboardButtonRow(btnStart, btnAbout)
		btnRate := tgbotapi.NewKeyboardButton("Please give us five Stars!")
		btnRow2 := tgbotapi.NewKeyboardButtonRow(btnRate)
		keyboard := tgbotapi.NewReplyKeyboard(btnRow1, btnRow2)

		keyboard.OneTimeKeyboard = true

		message.ReplyMarkup = keyboard
		_, err := u.Send(message)
		if err != nil {
			u.Println("Error sending Init keyboard:" + err.Error())
		}

	case ConfigState:

		btnHost := tgbotapi.NewKeyboardButton("Host")
		btnPort := tgbotapi.NewKeyboardButton("Port")
		btnRow1 := tgbotapi.NewKeyboardButtonRow(btnHost, btnPort)
		btnUser := tgbotapi.NewKeyboardButton("User")
		btnPass := tgbotapi.NewKeyboardButton("Pass")
		btnRow2 := tgbotapi.NewKeyboardButtonRow(btnUser, btnPass)
		btnReady := tgbotapi.NewKeyboardButton("Start sending commands!")
		btnRow3 := tgbotapi.NewKeyboardButtonRow(btnReady)
		keyboard := tgbotapi.NewReplyKeyboard(btnRow1, btnRow2, btnRow3)

		keyboard.OneTimeKeyboard = true

		message.ReplyMarkup = keyboard
		_, err := u.Send(message)
		if err != nil {
			u.Println("Error sending Config keyboard:" + err.Error())
		}

	case ReadyState:
	}
}

func SendRateInline(u *User) {

	u.Println("It's an opensource project, help us!")

	u.Println("Please access the link below and give us five Stars!")

	message := tgbotapi.NewMessage(u.ChatID, "https://telegram.me/storebot?start=sshclientbot")
	message.DisableWebPagePreview = true

	_, err := u.Send(message)
	if err != nil {
		u.Println("Error sending Rate keyboard: " + err.Error())
	}

}

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/crypto/ssh"
)

const tokenEnv string = "botToken"
const urlEnv string = "botUrl"
const portEnv string = "PORT"

type User struct {
	ChatID    int64
	Connected bool
	Host      string
	User      string
	Pass      string
	Session   *ssh.Session
	bot       *tg.BotAPI
}

// int = User.ID
var users = map[int]*User{}

func main() {

	token := os.Getenv(tokenEnv)
	if token == "" {
		log.Panic("TOKEN ENV NOT FOUND!")
	}

	url := os.Getenv(urlEnv)
	if url == "" {
		log.Panic("URL ENV NOT FOUND!")
	}

	port := os.Getenv(portEnv)
	if port == "" {
		log.Panic("PORT ENV NOT FOUND!")
	}

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	/*
		_, err = bot.SetWebhook(tg.NewWebhook("https://" + url + "/")) // +bot.Token, "cert.pem"
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Webhook to %s\n", "https://"+url+"/") // +bot.Token

		log.Printf("Serving to %s\n", ":"+port+"/") // +bot.Token

		updates := bot.ListenForWebhook("0.0.0.0:" + port + "/") //  +"/"+bot.Token

	*/

	//go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	u := tg.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		var msg tg.MessageConfig

		switch update.Message.Text {

		case "/start":
			bot.Send(tg.NewMessage(update.Message.Chat.ID, "Hello "+update.Message.From.UserName+"!"))
			continue

		case "/help":
			bot.Send(tg.NewMessage(update.Message.Chat.ID, "To connect use /ssh command"))
			continue

		case "/ssh":
			users[update.Message.From.ID] = &User{bot: bot, ChatID: update.Message.Chat.ID}
			bot.Send(tg.NewMessage(update.Message.Chat.ID, "Host:"))
			continue

		}

		//
		// User restrict methods
		//
		u, ok := users[update.Message.From.ID]

		// Block if user not in session
		if !ok {
			msg = tg.NewMessage(update.Message.Chat.ID, "Type /help for some instructions or /ssh to open a connection")
			continue
		}

		//
		// Here we have an User
		//
		if u.Host == "" {
			u.Host = update.Message.Text
			u.message("User:")
			continue
		}

		if u.User == "" {
			u.User = update.Message.Text
			u.message("Pass:")
			continue
		}

		if u.Pass == "" {
			u.Pass = update.Message.Text
			// Do not break, it goes start a session
		}

		if u.Session == nil {

			sshConfig := &ssh.ClientConfig{
				User: u.User,
				Auth: []ssh.AuthMethod{
					ssh.Password(u.Pass),
				},
			}

			client, err := ssh.Dial("tcp", u.Host+":22", sshConfig)
			if err != nil {
				u.message("Failed to dial:" + err.Error())
				// Clear configs to request it again
				u.clear()
				continue
			}

			session, err := client.NewSession()
			if err != nil {
				u.message("Failed to create session: " + err.Error())
				// Clear configs to request it again
				u.clear()
				continue
			}

			u.Session = session
			defer session.Close()

			err = u.listen()
			if err != nil {
				u.message("Failed to create Pipe: " + err.Error())
				u.clear()
				continue
			}

			u.message("SSH Session is open")
			continue
		}

		//
		// Here we a session
		//

		// Sesssion commands
		switch update.Message.Text {
		case "/exit":
			err = u.exit()
			if err != nil {
				u.message("Error closing sesion: " + err.Error())
				continue
			}
			u.message("Sesion close")
			continue

		case "/kill":
			err = u.Session.Signal(ssh.SIGKILL)
			if err != nil {
				u.message("Error sending KILL signal: " + err.Error())
				continue
			}
			u.message("Kill sent")
			continue
		}

		var b bytes.Buffer
		u.Session.Stdout = &b

		var e bytes.Buffer
		u.Session.Stderr = &e

		err = u.Session.Run(update.Message.Text)
		if err != nil {
			u.message("Error running the command:" + err.Error())
			u.clear()
			continue
		}

		u.listen()

		u.message("RUN: " + b.String())

		bot.Send(msg)
	}

}

func (u *User) message(line string) {
	u.bot.Send(tg.NewMessage(u.ChatID, line))
}

func (u *User) send(msg tg.MessageConfig) {
	u.bot.Send(msg)
}

func (u *User) listen() error {

	stdout, err := u.Session.StdoutPipe()
	if err != nil {
		return err
	}
	go u.read(stdout, "OUT")

	stderr, err := u.Session.StderrPipe()
	if err != nil {
		return err
	}
	go u.read(stderr, "ERR")

	return nil
}

func (u *User) read(rd io.Reader, sulf string) {
	log.Println("*** READING SOMETHING!")
	bio := bufio.NewReader(rd)
	line := ""
	var err error
	for err != io.EOF {
		str, hasMoreInLine, err := bio.ReadLine()
		if err != nil && err != io.EOF {
			fmt.Println("### " + err.Error())
			break
		}

		line = line + string(str)

		u.message(sulf + ": " + line)
		line = ""
		if err == io.EOF {
			break
		}
		if hasMoreInLine {
			continue // reading more
		}
	}
}

func (u *User) exit() error {
	if u.Session != nil {
		return u.Session.Close()
	}
	return nil
}

func (u *User) clear() {
	u.Connected = false
	u.Host = ""
	u.User = ""
	u.Pass = ""
	u.Session = nil
}

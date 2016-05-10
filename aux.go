package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/braintree/manners"
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

func StartWebServer() {

	ip := os.Getenv(ipEnv)
	if ip == "" {
		log.Panic("PORT ENV NOT FOUND!")
	}

	port := os.Getenv(portEnv)
	if port == "" {
		log.Panic("PORT ENV NOT FOUND!")
	}
	// WebServer Workaround
	// These routes are pingged for services just to never idles this instance
	mux := http.NewServeMux()
	// Starting a Web Server never idles our instance
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<a href=\"https://www.uptimia.com/\" target=\"_blank\"><img src=\"https://www.uptimia.com/status?hash=54bd14d2474753185fb6a66b9239a3f8\" width=\"130\" height=\"auto\" alt=\"Website monitoring | Uptimia\" title=\"Website monitoring | Uptimia\"></a>"))
	})
	// Shut the server down gracefully
	processStopedBySignal()
	// Manners allows you to shut your Go webserver down gracefully, without dropping any requests
	err := manners.ListenAndServe(ip+":"+port, mux)
	if err != nil {
		log.Panic(err)
		return
	} else {
		log.Println("Server listening at " + ip + ":" + port)
	}
	defer manners.Close()
	// END of WebServer Workaround
}

// Shut the server down gracefully if receive a interrupt signal
func processStopedBySignal() {
	// Stop server if someone kills the process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	//signal.Notify(c, syscall.SIGSTOP)
	signal.Notify(c, syscall.SIGABRT) // ABORT
	signal.Notify(c, syscall.SIGKILL) // KILL
	signal.Notify(c, syscall.SIGTERM) // TERMINATION
	signal.Notify(c, syscall.SIGINT)  // TERMINAL INTERRUPT (Ctrl+C)
	signal.Notify(c, syscall.SIGSTOP) // STOP
	signal.Notify(c, syscall.SIGTSTP) // TERMINAL STOP (Ctrl+Z)
	signal.Notify(c, syscall.SIGQUIT) // QUIT (Ctrl+\)
	go func() {
		fmt.Println("THIS PROCESS IS WAITING SIGNAL TO STOP GRACEFULLY")
		for sig := range c {
			fmt.Println("\n\nSTOPED BY SIGNAL:", sig.String())
			fmt.Println("SHUTTING DOWN GRACEFULLY!")
			fmt.Println("\nGod bye!")
			manners.Close()
			os.Exit(1)
		}
	}()
}

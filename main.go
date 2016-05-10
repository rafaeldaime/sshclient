package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/braintree/manners"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const tokenEnv string = "TOKEN"
const ipEnv string = "OPENSHIFT_GO_IP"
const portEnv string = "OPENSHIFT_GO_PORT"

type State uint

const (
	InitState   = 0
	ConfigState = 1
	ReadyState  = 2
)

func main() {

	token := os.Getenv(tokenEnv)
	if token == "" {
		log.Panic("TOKEN ENV NOT FOUND!")
	}

	ip := os.Getenv(ipEnv)
	if ip == "" {
		log.Panic("PORT ENV NOT FOUND!")
	}

	port := os.Getenv(portEnv)
	if port == "" {
		log.Panic("PORT ENV NOT FOUND!")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	news := tgbotapi.NewUpdate(0)
	news.Timeout = 60

	updates, err := bot.GetUpdatesChan(news)

	mux := http.NewServeMux()
	// Starting a Web Server never idles our instance
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ON!"))
	})
	http.Handle("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<a href=\"https://www.uptimia.com/\" target=\"_blank\"><img src=\"https://www.uptimia.com/status?hash=54bd14d2474753185fb6a66b9239a3f8\" width=\"130\" height=\"auto\" alt=\"Website monitoring | Uptimia\" title=\"Website monitoring | Uptimia\"></a>"))
	})
	// Shut the server down gracefully
	processStopedBySignal()

	// Manners allows you to shut your Go webserver down gracefully, without dropping any requests
	err = manners.ListenAndServe(ip+":"+port, mux)
	if err != nil {
		log.Panic(err)
		return
	} else {
		log.Println("Server listening at " + ip + ":" + port)
	}
	defer manners.Close()
	// END Web Server

	for update := range updates {

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// This bot doesn't answare for inline queries
		if update.InlineQuery != nil {
			continue
		}

		u := GetUser(update)

		// Block if user not in session
		if u == nil {
			u = NewUser(bot, update)
			u.State = InitState
		}

		// Handle the actual command
		go Handle(u, update)

	}

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

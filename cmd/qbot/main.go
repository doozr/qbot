package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/doozr/guac"
	"github.com/doozr/jot"
	"github.com/doozr/qbot"
	"github.com/doozr/qbot/command"
	"github.com/doozr/qbot/queue"
	"github.com/doozr/qbot/usercache"
)

func main() {
	log.Printf("Qbot version %s", qbot.Version)

	// Turn on jot if required
	if os.Getenv("QBOT_DEBUG") == "true" {
		jot.Enable()
	}

	token, filename := parseCLI()

	waitGroup := sync.WaitGroup{}
	done := make(qbot.DoneChan)

	q := loadQueueOrDie(filename)

	client := connectToSlackOrDie(token)

	userCache := getUserListOrDie(client)

	commands := command.New(client.Name(), userCache)

	notify := qbot.CreateNotifier(client.IMOpen, client.PostMessage)
	persist := qbot.CreatePersister(ioutil.WriteFile, filename, q)
	userChangeHandler := qbot.CreateUserChangeHandler(userCache)

	publicHandler := qbot.CreateMessageHandler(q, qbot.PublicCommands(commands), notify, persist)
	privateHandler := qbot.CreateMessageHandler(q, qbot.PrivateCommands(commands), notify, persist)
	messageHandler := qbot.CreateMessageDirector(client.ID(), client.Name(), publicHandler, privateHandler)

	receiver := qbot.CreateEventReceiver(client)
	events := qbot.Receive(receiver, done, &waitGroup)

	qbot.StartKeepAlive(client.Ping, time.After, done, &waitGroup)

	log.Print("Ready")
	dispatcher := qbot.CreateDispatcher(1*time.Minute, messageHandler, userChangeHandler)
	abort := qbot.Dispatch(dispatcher, events, done, &waitGroup)
	sig := addSignalHandler()
	wait(sig, abort)

	close(done)
	client.Close()
	waitGroup.Wait()

	jot.Print("qbot: shutdown complete")
}

func parseCLI() (token, filename string) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: qbot <token> <data file>")
		os.Exit(1)
	}
	token = os.Args[1]
	filename = os.Args[2]
	return
}

func connectToSlackOrDie(token string) guac.RealTimeClient {
	client, err := guac.New(token).RealTime()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected to slack as ", client.Name())
	return client
}

func loadQueueOrDie(filename string) (q queue.Queue) {
	q = queue.Queue{}
	if _, err := os.Stat(filename); err == nil {
		dat, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error loading queue: %s", err)
		}
		json.Unmarshal(dat, &q)
		jot.Printf("loadQueue: read queue from %s: %v", filename, q)
		log.Printf("Loaded queue from %s", filename)
	}
	return q
}

func getUserListOrDie(client guac.WebClient) (userCache usercache.UserCache) {
	log.Println("Getting user list")
	users, err := client.UsersList()
	if err != nil {
		log.Fatal(err)
	}
	userCache = usercache.New(users)
	jot.Printf("loaded user list with %d users", userCache.Count())
	return
}

func addSignalHandler() chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGKILL)
	return sig
}

func wait(sig chan os.Signal, abort chan error) {
	select {
	case err := <-abort:
		if err != nil {
			log.Print("Error: ", err)
		}
		log.Print("Execution terminated - shutting down")
	case s := <-sig:
		log.Printf("Received %s signal - shutting down", s)
	}
}
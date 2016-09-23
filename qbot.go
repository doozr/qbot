package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/doozr/guac"
	"github.com/doozr/jot"
	"github.com/doozr/qbot/command"
	"github.com/doozr/qbot/dispatch"
	"github.com/doozr/qbot/notification"
	"github.com/doozr/qbot/queue"
	"github.com/doozr/qbot/usercache"
)

// Version is the current release version
var Version string

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: qbot <token> <data file>")
		os.Exit(1)
	}

	if Version != "" {
		log.Printf("Qbot version %s", Version)
	} else {
		log.Printf("Qbot <unversioned build>")
	}

	// Get command line parameters
	token := os.Args[1]
	filename := os.Args[2]

	// Turn on jot if required
	if os.Getenv("QBOT_DEBUG") == "true" {
		jot.Enable()
	}

	// Synchronisation primitives
	waitGroup := sync.WaitGroup{}
	done := make(chan struct{})

	// Connect to Slack
	client, err := guac.New(token).RealTime()
	if err != nil {
		log.Fatal("Error connecting to Slack ", err)
	}
	log.Print("Connected to slack as ", client.Name())

	// Instantiate state
	userCache := getUserList(client)
	name := client.Name()
	jot.Print("qbot: name is ", name)
	q := loadQueue(filename)

	// Set up command and response processors
	notifications := notification.New(userCache)
	commands := command.New(notifications, userCache)

	// Create dispatchers
	notify := dispatch.NewNotifier(client)
	persist := dispatch.NewPersister(filename)
	messageHandler := dispatch.NewMessageHandler(client.ID(), client.Name(), q, commands, notify, persist)
	userChangeHandler := dispatch.NewUserChangeHandler(userCache)

	// keepalive
	waitGroup.Add(1)
	go keepalive(client, done, &waitGroup)

	// Dispatch incoming events
	jot.Println("qbot: ready to receive events")
	dispatcher := dispatch.New(messageHandler, userChangeHandler)
	abort := dispatcher.Listen(client, 2*time.Minute, done, &waitGroup)

	// Wait for signals to stop
	sig := addSignalHandler()

	// Wait for a signal
	select {
	case err = <-abort:
		if err != nil {
			log.Print("Error: ", err)
		}
		log.Print("Execution terminated - shutting down")
	case s := <-sig:
		log.Printf("Received %s signal - shutting down", s)
	}

	jot.Print("qbot: closing done channel")
	close(done)

	jot.Print("qbot: closing connection")
	client.Close()

	jot.Print("qbot: waiting for dispatch to terminate")
	waitGroup.Wait()

	jot.Print("qbot: shutdown complete")
}

func addSignalHandler() chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGKILL)
	return sig
}

func getUserList(client guac.WebClient) (userCache *usercache.UserCache) {
	log.Println("Getting user list")
	users, err := client.UsersList()
	if err != nil {
		log.Fatal(err)
	}
	userCache = usercache.New(users)
	jot.Print("loaded user list: ", userCache)
	return
}

func loadQueue(filename string) (q queue.Queue) {
	q, err := queue.Load(filename)
	if err != nil {
		log.Fatalf("Error loading queue: %s", err)
	}
	log.Printf("Loaded queue from %s", filename)
	return
}

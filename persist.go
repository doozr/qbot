package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/doozr/jot"
	"github.com/doozr/qbot/queue"
)

// Persister handles exporting the queue to persistent media
type Persister func(queue.Queue) error

// NewPersister creates a new Persister
func createPersister(filename string) Persister {
	var oldQ queue.Queue

	return func(q queue.Queue) (err error) {
		jot.Print("persist: queue to save ", q)
		if oldQ.Equal(q) {
			return
		}

		j, err := json.Marshal(q)
		if err != nil {
			log.Print("Error serialising qeuue: ", err)
			return
		}

		jot.Print("queue: writing queue JSON: ", string(j))
		err = ioutil.WriteFile(filename, j, 0644)
		if err != nil {
			log.Printf("Error saving file to %s: %s", filename, err)
			return
		}

		oldQ = q
		jot.Print("persist: saved to ", filename)
		return
	}
}

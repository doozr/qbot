package command_test

import (
	"testing"

	"github.com/doozr/qbot/queue"
)

func TestDone(t *testing.T) {
	testCommand(t, command.Done, []CommandTest{
		{
			test:             "drop token",
			startQueue:       queue.Queue([]queue.Item{{"U123", "Banana"}}),
			channel:          "C1A2B3C",
			user:             "U123",
			expectedQueue:    queue.Queue{},
			expectedResponse: "<@U123|craig> (Banana) has finished with the token\nThe token is up for grabs",
		},
		{
			test:             "drop token and give it to the next in line",
			startQueue:       queue.Queue([]queue.Item{{"U123", "Banana"}, {"U456", "Next up"}}),
			channel:          "C1A2B3C",
			user:             "U123",
			expectedQueue:    queue.Queue([]queue.Item{{"U456", "Next up"}}),
			expectedResponse: "<@U123|craig> (Banana) has finished with the token\n*<@U456|edward> (Next up) now has the token*",
		},
		{
			test:             "warns if user does not have the token",
			startQueue:       queue.Queue([]queue.Item{{"U123", "Banana"}, {"U456", "Next up"}}),
			channel:          "C1A2B3C",
			user:             "U456",
			expectedQueue:    queue.Queue([]queue.Item{{"U123", "Banana"}, {"U456", "Next up"}}),
			expectedResponse: "<@U456|edward> You cannot be done if you don't have the token",
		},
		{
			test:             "does nothing if the queue is empty",
			startQueue:       queue.Queue{},
			channel:          "C1A2B3C",
			user:             "U456",
			expectedQueue:    queue.Queue{},
			expectedResponse: "",
		},
	})
}

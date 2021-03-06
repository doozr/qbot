package command_test

import (
	"testing"

	"github.com/doozr/guac"
	. "github.com/doozr/qbot/command"
	"github.com/doozr/qbot/queue"
	"github.com/doozr/qbot/usercache"
)

type CommandTest struct {
	test             string
	startQueue       queue.Queue
	channel          string
	user             string
	args             string
	expectedQueue    queue.Queue
	expectedResponse string
}

var id = "U12345"
var name = "the_bot_name"

var userCache = usercache.New([]guac.UserInfo{
	{
		ID:   "U123",
		Name: "craig",
	},
	{
		ID:   "U456",
		Name: "edward",
	},
	{
		ID:   "U789",
		Name: "andrew",
	},
	{
		ID:   id,
		Name: name,
	},
})

func testCommand(t *testing.T, fn Command, tests []CommandTest) {
	for _, tt := range tests {
		q, r := fn(tt.startQueue, tt.channel, tt.user, tt.args)
		assertQueue(t, tt.test, tt.expectedQueue, q)
		assertResponse(t, tt.test, tt.channel, tt.expectedResponse, r)
	}
}

func assertQueue(t *testing.T, test string, expected, actual queue.Queue) {
	if !actual.Equal(expected) {
		t.Errorf("%s: expected queue '%v', got '%v'", test, expected, actual)
	}
}

func assertResponse(t *testing.T, test, channel, message string, actual Notification) {
	expected := Notification{
		Channel: channel,
		Message: message,
	}
	if actual != expected {
		t.Errorf("%s: expected response '%v', got '%v'", test, expected, actual)
	}
}

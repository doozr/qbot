package queue_test

import (
	"testing"

	. "github.com/doozr/qbot/queue"
	"github.com/stretchr/testify/assert"
)

var John = Item{ID: "john", Reason: "done some coding"}
var Jimmy = Item{ID: "jimmy", Reason: "fix some bugs"}
var Mick = Item{ID: "mick", Reason: "refactoring"}
var Colin = Item{ID: "colin", Reason: "adding bugs"}

func TestEqual(t *testing.T) {
	q := Queue{John, Jimmy, Mick}
	other := Queue{John, Jimmy, Mick}

	assert.Equal(t, true, q.Equal(other))
}

func TestUnequalIfDifferentLengths(t *testing.T) {
	q := Queue{John, Jimmy, Mick}
	other := Queue{John, Jimmy}

	assert.Equal(t, false, q.Equal(other))
}

func TestUnequalIfDifferentContent(t *testing.T) {
	q := Queue{John, Jimmy, Mick}
	other := Queue{John, Jimmy, Colin}

	assert.Equal(t, false, q.Equal(other))
}

func TestCreateQueue(t *testing.T) {
	q := Queue{}
	assert.Equal(t, 0, len(q))
}

func TestActiveIsEmptyIfQueueIsEmpty(t *testing.T) {
	q := Queue{}
	assert.Equal(t, Item{}, q.Active())
}

func TestCreateQueueWithEntries(t *testing.T) {
	q := Queue{Mick, John}
	assert.Equal(t, 2, len(q))
}

func TestAddImmutable(t *testing.T) {
	q := Queue{}
	q.Add(Mick)
	assert.Equal(t, 0, len(q))
}

func TestAdd(t *testing.T) {
	q := Queue{}
	q = q.Add(Mick)
	assert.Equal(t, 1, len(q))

	active := q.Active()
	assert.Equal(t, Mick, active)
}

func TestAddDuplicate(t *testing.T) {
	q := Queue{Mick}
	q = q.Add(Mick)
	assert.Equal(t, 1, len(q))
}

func TestAddWithDifferentReason(t *testing.T) {
	q := Queue{Mick}
	q = q.Add(Item{ID: "mick", Reason: "wrote tests"})
	assert.Equal(t, 2, len(q))
}

func TestWaitingWhenEmpty(t *testing.T) {
	q := Queue{}
	w := q.Waiting()
	assert.Equal(t, 0, len(w))
}

func TestWaitingWhenOnlyOne(t *testing.T) {
	q := Queue{Mick}
	w := q.Waiting()
	assert.Equal(t, 0, len(w))
}

func TestWaitingWhenMoreThanOne(t *testing.T) {
	q := Queue{Mick, John, Jimmy}
	expected := []Item{John, Jimmy}
	assert.Equal(t, expected, q.Waiting())
}

func TestRemoveImmutable(t *testing.T) {
	q := Queue{Mick}
	q.Remove(Mick)
	assert.Equal(t, 1, len(q))
}

func TestRemoveActive(t *testing.T) {
	q := Queue{Mick, John}
	q = q.Remove(Mick)
	assert.Equal(t, 1, len(q))
	assert.Equal(t, John, q.Active())
}

func TestRemoveMiddle(t *testing.T) {
	q := Queue{Mick, Jimmy, John}
	q = q.Remove(Jimmy)
	assert.Equal(t, 2, len(q))
	assert.NotContains(t, q, Jimmy)
}

func TestRemoveLast(t *testing.T) {
	q := Queue{Mick, John}
	q = q.Remove(John)
	assert.Equal(t, 1, len(q))
	assert.NotContains(t, q, John)
}

func TestRemoveNotPresent(t *testing.T) {
	q := Queue{Mick, John}
	q = q.Remove(Jimmy)
	assert.Equal(t, 2, len(q))
}

func TestYield(t *testing.T) {
	q := Queue{Mick, John}
	q = q.Yield()
	assert.Equal(t, John, q.Active())
}

func TestYieldAlone(t *testing.T) {
	q := Queue{Mick}
	q = q.Yield()
	assert.Equal(t, Mick, q.Active())
}

func TestBargeWhenEmpty(t *testing.T) {
	q := Queue{}
	q = q.Barge(Mick)
	assert.Equal(t, Mick, q.Active())
}

func TestBargeWhenActive(t *testing.T) {
	q := Queue{Mick, John}
	q = q.Barge(Mick)
	assert.Equal(t, Mick, q.Active())
}

func TestBargeWhenOnlyOne(t *testing.T) {
	q := Queue{John}
	q = q.Barge(Mick)
	expected := []Item{Mick}
	assert.Equal(t, expected, q.Waiting())
}

func TestBargeWhenOnlyTwoAndAlreadySecond(t *testing.T) {
	q := Queue{John, Mick}
	q = q.Barge(Mick)
	expected := []Item{Mick}
	assert.Equal(t, expected, q.Waiting())
}

func TestBargeWhenOnlyTwo(t *testing.T) {
	q := Queue{John, Jimmy}
	q = q.Barge(Mick)
	expected := []Item{Mick, Jimmy}
	assert.Equal(t, expected, q.Waiting())
}

func TestBargeWhenAlreadySecond(t *testing.T) {
	q := Queue{John, Mick}
	q = q.Barge(Mick)
	expected := []Item{Mick}
	assert.Equal(t, expected, q.Waiting())
}

func TestBargeWhenNotPresent(t *testing.T) {
	q := Queue{John, Colin, Jimmy}
	q = q.Barge(Mick)
	expected := []Item{Mick, Colin, Jimmy}
	assert.Equal(t, expected, q.Waiting())
}

func TestBargeWhenPresent(t *testing.T) {
	q := Queue{John, Colin, Mick, Jimmy}
	q = q.Barge(Mick)
	expected := []Item{Mick, Colin, Jimmy}
	assert.Equal(t, expected, q.Waiting())
}

func TestDelegateWhenPresent(t *testing.T) {
	q := Queue{John, Colin, Mick}
	q = q.Delegate(Colin, Jimmy)
	expected := []Item{Jimmy, Mick}
	assert.Equal(t, expected, q.Waiting())
}

func TestDelegateWhenActive(t *testing.T) {
	q := Queue{John, Colin, Mick}
	q = q.Delegate(John, Jimmy)
	expected := Jimmy
	assert.Equal(t, expected, q.Active())
}

func TestDelegateWhenNotPresent(t *testing.T) {
	q := Queue{John, Colin, Mick}
	q = q.Delegate(Jimmy, Colin)
	expected := []Item{Colin, Mick}
	assert.Equal(t, expected, q.Waiting())
}

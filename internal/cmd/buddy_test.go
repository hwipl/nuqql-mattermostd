package cmd

import (
	"testing"
)

func TestNewBuddy(t *testing.T) {
	user := "TestBuddyID"
	name := "TestBuddyName"
	status := "online"

	b := newBuddy(user, name, status)
	if b.user != user {
		t.Errorf("got %s, wanted %s", b.user, user)
	}
	if b.name != name {
		t.Errorf("got %s, wanted %s", b.name, name)
	}
	if b.status != status {
		t.Errorf("got %s, wanted %s", b.status, status)
	}
}

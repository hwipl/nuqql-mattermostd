package cmd

import (
	"testing"
)

func TestSplitAccountUser(t *testing.T) {
	a := account{
		User: "testuser@testserver.com:8065",
	}
	username, server := a.splitAccountUser()

	// check username
	want := "testuser"
	got := username
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}

	// check server address
	want = "testserver.com:8065"
	got = server
	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

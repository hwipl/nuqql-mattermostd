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

func TestAccountStart(t *testing.T) {
	// test dummy account
	a := account{}
	a.start()
}

func TestAccountStop(t *testing.T) {
	// test dummy account
	a := account{}
	a.stop()
}

func TestGetFreeAccountID(t *testing.T) {
	accounts = make(map[int]*account)
	defer func() {
		// cleanup
		accounts = make(map[int]*account)
	}()

	// test empty
	want := 0
	got := getFreeAccountID()
	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// test filled
	accounts[0] = &account{ID: 0}
	accounts[1] = &account{ID: 1}
	accounts[2] = &account{ID: 2}

	want = 3
	got = getFreeAccountID()
	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// test filled (see above) with gap
	accounts[4] = &account{ID: 4}

	want = 3
	got = getFreeAccountID()
	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}

	// test with filled gap
	accounts[3] = &account{ID: 3}

	want = 5
	got = getFreeAccountID()
	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}
}

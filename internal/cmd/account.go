package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	// accountsFile is the json file that contains all accounts
	accountsFile = "accounts.json"

	// accounts contains all active accounts
	accounts = make(map[int]*account)
)

// account stores account information
type account struct {
	ID       int
	Protocol string
	User     string
	Password string

	client *mattermost
}

// splitAccountUser extracts username and server address from the account user
func (a *account) splitAccountUser() (username, server string) {
	username = strings.Split(a.User, "@")[0]
	server = strings.Split(a.User, "@")[1]
	return
}

// start starts the client for this account
func (a *account) start(ctx context.Context) {
	// skip non-mattermost accounts
	if a.Protocol != "mattermost" {
		return
	}

	// extract server and username from account user
	user, server := a.splitAccountUser()

	// start client
	logInfo("Starting account", a.ID)
	a.client = newClient(conf, a.ID, server, user, a.Password)
	go a.client.run(ctx)
}

// stop shuts down the client for this account
func (a *account) stop() {
	if a.client != nil {
		logInfo("Stopping account", a.ID)
		a.client.stop()
	}
}

// getAccount returns account with account ID
func getAccount(id int) *account {
	return accounts[id]
}

// getAccounts returns all accounts sorted by account ID
func getAccounts() []*account {
	// sort account ids
	ids := make([]int, len(accounts))
	i := 0
	for id := range accounts {
		ids[i] = id
		i++
	}
	sort.Ints(ids)

	// construct sorted slice of accounts
	accs := make([]*account, len(accounts))
	for i, id := range ids {
		accs[i] = accounts[id]
	}
	return accs
}

// getFreeAccountID returns the first free account ID
func getFreeAccountID() int {
	for i := 0; i < len(accounts); i++ {
		if accounts[i] == nil {
			return i
		}
	}
	return len(accounts)
}

// addAccount adds a new account with protocol, user and password and returns
// the new account's ID
func addAccount(ctx context.Context, protocol, user, password string) int {
	a := account{
		ID:       getFreeAccountID(),
		Protocol: protocol,
		User:     user,
		Password: password,
	}
	accounts[a.ID] = &a
	writeAccountsToFile()
	a.start(ctx)
	return a.ID
}

// delAccount removes the existing account with id
func delAccount(id int) bool {
	if accounts[id] != nil {
		accounts[id].stop()
		delete(accounts, id)
		writeAccountsToFile()
		return true
	}
	return false
}

// readAccountsFromFile reads accounts from file
func readAccountsFromFile() {
	file := filepath.Join(conf.Dir, accountsFile)

	// open file for reading
	f, err := os.Open(file)
	if err != nil {
		logError(err)
		return
	}

	// read accounts from file
	dec := json.NewDecoder(f)
	for {
		var a account
		err := dec.Decode(&a)
		if err == io.EOF {
			break
		}
		if err != nil {
			logFatal(err)
		}

		accounts[a.ID] = &a
	}
}

// writeAccountsToFile writes all accounts to file
func writeAccountsToFile() {
	file := filepath.Join(conf.Dir, accountsFile)

	// open file for writing
	f, err := os.Create(file)
	if err != nil {
		logFatal(err)
	}

	// make sure file is only readable and writable by the current user
	err = os.Chmod(file, 0600)
	if err != nil {
		logFatal(err)
	}

	// write accounts to file
	enc := json.NewEncoder(f)
	for _, a := range accounts {
		err := enc.Encode(&a)
		if err != nil {
			logFatal(err)
		}
	}
}

// startAccounts initializes all accounts and starts their clients
func startAccounts(ctx context.Context) {
	// read accounts
	readAccountsFromFile()
	for _, a := range accounts {
		a.start(ctx)
	}
}

// stopAccounts stops all accounts and their clients
func stopAccounts() {
	for _, a := range accounts {
		a.stop()
	}
}

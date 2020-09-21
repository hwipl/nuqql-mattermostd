package cmd

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sort"
)

var (
	// accounts contains all active accounts
	accounts = make(map[int]*account)
)

// account stores account information
type account struct {
	ID       int
	Protocol string
	User     string
	Password string
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

// readAccountsFromFile reads accounts from file
func readAccountsFromFile(file string) {
	// open file for reading
	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
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
			log.Fatal(err)
		}

		accounts[a.ID] = &a
	}
}

// writeAccountsToFile writes all accounts to file
func writeAccountsToFile(file string) {
	// open file for writing
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}

	// write accounts to file
	enc := json.NewEncoder(f)
	for _, a := range accounts {
		err := enc.Encode(&a)
		if err != nil {
			log.Fatal(err)
		}
	}
}

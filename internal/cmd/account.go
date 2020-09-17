package cmd

import (
	"encoding/json"
	"io"
	"log"
	"os"
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

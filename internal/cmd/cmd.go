package cmd

import (
	"log"
	"os"
)

// Run is the main entry point
func Run() {
	// make sure working directory exists
	err := os.MkdirAll(conf.dir, 0700)
	if err != nil {
		log.Fatal(err)
	}

	// start accounts and client connections
	startAccounts()

	// start server
	runServer()

	// stop all accounts and client connections
	stopAccounts()
}

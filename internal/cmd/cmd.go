package cmd

import (
	"log"
	"os"
)

// initDirectory makes sure the working directory exists
func initDirectory() {
	err := os.MkdirAll(conf.dir, 0700)
	if err != nil {
		log.Fatal(err)
	}

}

// Run is the main entry point
func Run() {
	// make sure working directory exists
	initDirectory()

	// start accounts and client connections
	startAccounts()

	// start server
	runServer()

	// stop all accounts and client connections
	stopAccounts()
}

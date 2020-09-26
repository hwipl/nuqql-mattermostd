package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// parseCommandLine parses the command line arguments
func parseCommandLine() {
	// configure command line arguments
	version := flag.Bool("v", false, "show version and exit")
	flag.StringVar(&conf.af, "af", conf.af, "set socket address "+
		"`family`: \"inet\" for AF_INET, \"unix\" for AF_UNIX")

	// parse command line arguments
	flag.Parse()

	// handle version command line argument
	if *version {
		fmt.Println("0.0dev")
		os.Exit(0)
	}
}

// initDirectory makes sure the working directory exists
func initDirectory() {
	err := os.MkdirAll(conf.dir, 0700)
	if err != nil {
		log.Fatal(err)
	}

}

// Run is the main entry point
func Run() {
	// parse command line arguments
	parseCommandLine()

	// make sure working directory exists
	initDirectory()

	// start accounts and client connections
	startAccounts()

	// start server
	runServer()

	// stop all accounts and client connections
	stopAccounts()
}

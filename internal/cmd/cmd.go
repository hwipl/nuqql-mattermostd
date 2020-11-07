package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	backendVersion = "0.1.0"
)

// parseCommandLine parses the command line arguments
func parseCommandLine() {
	// configure command line arguments
	version := flag.Bool("v", false, "show version and exit")
	flag.StringVar(&conf.AF, "af", conf.AF, "set socket address "+
		"`family`: \"inet\" for AF_INET, \"unix\" for AF_UNIX")
	flag.StringVar(&conf.Address, "address", conf.Address,
		"set AF_INET listen `address`")
	port := flag.Uint("port", uint(conf.Port), "set AF_INET listen `port`")
	flag.StringVar(&conf.Sockfile, "sockfile", conf.Sockfile,
		"set AF_UNIX socket `file` in working directory")
	flag.StringVar(&conf.Dir, "dir", conf.Dir, "set working `directory`")
	loglevel := flag.String("loglevel", conf.Loglevel,
		"set logging `level`: debug, info, warn, error")
	flag.BoolVar(&conf.DisableHistory, "disable-history",
		conf.DisableHistory, "disable message history")
	flag.BoolVar(&conf.PushAccounts, "push-accounts", conf.PushAccounts,
		"push accounts to client")
	flag.BoolVar(&conf.DisableFilterOwn, "disable-filterown",
		conf.DisableFilterOwn, "disable filtering of own messages")
	flag.BoolVar(&conf.DisableEncryption, "disable-encryption",
		conf.DisableEncryption, "disable TLS encryption")

	// parse command line arguments
	flag.Parse()

	// handle version command line argument
	if *version {
		fmt.Println(backendVersion)
		os.Exit(0)
	}

	// parse port number
	if *port > 65535 {
		log.Fatal("error parsing port ", *port)
	}
	conf.Port = uint16(*port)

	// handle log level
	if *loglevel != "" {
		conf.Loglevel = *loglevel
	}
}

// initDirectory makes sure the working directory exists
func initDirectory() {
	err := os.MkdirAll(conf.Dir, 0700)
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

	// initialize logging
	initLogging()

	// start client queue
	initClientQueue()

	// start accounts and client connections
	startAccounts()

	// start server
	runServer()

	// stop all accounts and client connections
	stopAccounts()

	// stop logging
	stopLogging()
}

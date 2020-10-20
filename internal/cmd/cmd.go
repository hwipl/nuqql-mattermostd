package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	backendVersion = "0.0.0-devel"
)

// parseCommandLine parses the command line arguments
func parseCommandLine() {
	// configure command line arguments
	version := flag.Bool("v", false, "show version and exit")
	flag.StringVar(&conf.af, "af", conf.af, "set socket address "+
		"`family`: \"inet\" for AF_INET, \"unix\" for AF_UNIX")
	flag.StringVar(&conf.address, "address", conf.address,
		"set AF_INET listen `address`")
	port := flag.Uint("port", uint(conf.port), "set AF_INET listen `port`")
	flag.StringVar(&conf.sockfile, "sockfile", conf.sockfile,
		"set AF_UNIX socket `file` in working directory")
	flag.StringVar(&conf.dir, "dir", conf.dir, "set working `directory`")
	loglevel := flag.String("loglevel", conf.loglevel,
		"set logging `level`: debug, info, warn, error")
	flag.BoolVar(&conf.disableHistory, "disable-history",
		conf.disableHistory, "disable message history")
	flag.BoolVar(&conf.pushAccounts, "push-accounts", conf.pushAccounts,
		"push accounts to client")
	disableEncryption := flag.Bool("disable-encryption", false,
		"disable TLS encryption")
	flag.BoolVar(&conf.disableFilterOwn, "disable-filterown",
		conf.disableFilterOwn, "disable filtering of own messages")

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
	conf.port = uint16(*port)

	// handle log level
	if *loglevel != "" {
		conf.loglevel = *loglevel
	}

	// handle disable encryption
	if *disableEncryption {
		httpPrefix = "http://"
		webSocketPrefix = "ws://"
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

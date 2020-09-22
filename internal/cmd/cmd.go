package cmd

// Run is the main entry point
func Run() {
	// init accounts and client connections
	initAccounts()

	// start server
	runServer()
}

package cmd

// Run is the main entry point
func Run() {
	// start accounts and client connections
	startAccounts()

	// start server
	runServer()
}

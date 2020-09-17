package cmd

// Run is the main entry point
func Run() {
	readAccountsFromFile("accounts.json")
	runClient()
}

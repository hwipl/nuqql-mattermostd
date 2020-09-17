package cmd

import "strings"

// Run is the main entry point
func Run() {
	// read accounts
	readAccountsFromFile("accounts.json")
	for _, a := range accounts {
		// skip non-mattermost accounts
		if a.Protocol != "mattermost" {
			continue
		}

		// extract server and username from account user
		user := strings.Split(a.User, "@")[0]
		server := strings.Split(a.User, "@")[1]

		// start client
		go runClient(server, user, a.Password)
	}

	// wait
	select {}
}

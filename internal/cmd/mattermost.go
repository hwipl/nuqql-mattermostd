package cmd

import (
	"log"

	"github.com/mattermost/mattermost-server/v5/model"
)

// mattermost stores mattermost client information
type mattermost struct {
	server   string
	username string
	password string
	client   *model.Client4
}

// getErrorMessage converts an AppError to a string
func getErrorMessage(err *model.AppError) string {
	return err.Message + " " + err.Id + " " + err.DetailedError
}

// connect connects to a mattermost server
func (m *mattermost) connect() {
	m.client = model.NewAPIv4Client(m.server)

	// check is server is running
	props, resp := m.client.GetOldClientConfig("")
	if resp.Error != nil {
		log.Fatal(getErrorMessage(resp.Error))
	}
	log.Println("Server is running version", props["Version"])

	// login
	user, resp := m.client.Login(m.username, m.password)
	if resp.Error != nil {
		log.Fatal(getErrorMessage(resp.Error))
	}
	log.Println("Logged in as user", user.Username)
}

// runClient runs a mattermost client
func runClient() {
	// TODO: fix server address, add username and password
	m := mattermost{
		server:   "http://localhost:8065",
		username: "",
		password: "",
	}
	m.connect()
}

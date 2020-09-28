package cmd

import (
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

// mattermost stores mattermost client information
type mattermost struct {
	accountID int
	server    string
	username  string
	password  string
	client    *model.Client4
	user      *model.User
	websock   *model.WebSocketClient
	done      chan bool
}

// getErrorMessage converts an AppError to a string
func getErrorMessage(err *model.AppError) string {
	return err.Message + " " + err.Id + " " + err.DetailedError
}

// getBuddies returns a list of teams and channels the user is in
func (m *mattermost) getBuddies() []*buddy {
	var buddies []*buddy

	// get teams
	teams, resp := m.client.GetTeamsForUser(m.user.Id, "")
	if resp.Error != nil {
		log.Fatal(getErrorMessage(resp.Error))
	}
	for _, t := range teams {
		// get channels
		channels, resp := m.client.GetChannelsForTeamForUser(t.Id,
			m.user.Id, false, "")
		if resp.Error != nil {
			log.Fatal(getErrorMessage(resp.Error))
		}
		for _, c := range channels {
			user := c.Id
			name := c.DisplayName + " (" + t.DisplayName + ")"
			status := "GROUP_CHAT"
			b := newBuddy(user, name, status)
			buddies = append(buddies, b)
		}
	}

	return buddies
}

// sendMsg sends a message to channel
func (m *mattermost) sendMsg(channel string, msg string) {
	post := &model.Post{}
	post.ChannelId = channel
	post.Message = msg

	if _, resp := m.client.CreatePost(post); resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
	}
}

// handleWebSocketEvent handles events from the websocket
func (m *mattermost) handleWebSocketEvent(event *model.WebSocketEvent) {
	log.Println("WebSocket Event:", event.EventType())

	// only handle posted events
	if event.Event != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	// handle post
	post := model.PostFromJson(strings.NewReader(
		event.Data["post"].(string)))
	if post != nil {
		// filter own messages
		if post.UserId == m.user.Id {
			return
		}
		log.Println("Message:", post.CreateAt, post.ChannelId,
			post.UserId, post.Message)

		// get user who sent this message
		user, resp := m.client.GetUser(post.UserId, "")
		if resp.Error != nil {
			log.Fatal(getErrorMessage(resp.Error))
		}

		// construct message with format:
		// chat: msg: <acc_id> <chat> <timestamp> <sender> <message>
		// and send it via the client queue
		msg := fmt.Sprintf("chat: msg: %d %s %d %s %s\r\n",
			m.accountID, post.ChannelId, post.CreateAt/1000,
			user.Username, html.EscapeString(post.Message))
		clientQueue.send(msg)
	}
}

// connect connects to a mattermost server
func (m *mattermost) connect() {
	m.client = model.NewAPIv4Client("http://" + m.server)

	// check is server is running
	props, resp := m.client.GetOldClientConfig("")
	if resp.Error != nil {
		log.Fatal(getErrorMessage(resp.Error))
	}
	log.Println("Server is running version", props["Version"])

	// login
	m.user, resp = m.client.Login(m.username, m.password)
	if resp.Error != nil {
		log.Fatal(getErrorMessage(resp.Error))
	}
	log.Println("Logged in as user", m.user.Username)

	// get teams
	teams, resp := m.client.GetTeamsForUser(m.user.Id, "")
	if resp.Error != nil {
		log.Fatal(getErrorMessage(resp.Error))
	}
	for _, t := range teams {
		log.Println("User", m.user.Username, "is a member of team",
			t.Name, "("+t.DisplayName+")")

		// get channels
		channels, resp := m.client.GetChannelsForTeamForUser(t.Id,
			m.user.Id, false, "")
		if resp.Error != nil {
			log.Fatal(getErrorMessage(resp.Error))
		}
		for _, c := range channels {
			log.Println("User", m.user.Username, "is in team",
				t.Name+"'s channel", c.Name,
				"("+c.DisplayName+")")
		}
	}

	// create websocket and start listening for events
	websock, err := model.NewWebSocketClient4("ws://"+m.server,
		m.client.AuthToken)
	if err != nil {
		log.Fatal(getErrorMessage(err))
	}
	defer websock.Close()
	m.websock = websock
	m.websock.Listen()

	// handle websocket events
	for {
		select {
		case event := <-m.websock.EventChannel:
			m.handleWebSocketEvent(event)
		case <-m.done:
			return
		}
	}
}

// run starts the mattermost client
func (m *mattermost) run() {
	m.connect()
}

// stop shuts down the mattermost client
func (m *mattermost) stop() {
	m.done <- true
}

// newClient creates a new mattermost client
func newClient(accountID int, server, username, password string) *mattermost {
	m := mattermost{
		accountID: accountID,
		server:    server,
		username:  username,
		password:  password,
		done:      make(chan bool, 1),
	}
	return &m
}

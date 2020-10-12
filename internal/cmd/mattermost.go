package cmd

import (
	"fmt"
	"html"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
)

var (
	// filterOwn toggles filtering of own messages
	filterOwn = true

	// httpPrefix is prepended to the server to form a http url
	httpPrefix = "https://"
	// webSocketPrefix is prepended to the server to form a websocket url
	webSocketPrefix = "wss://"
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
	mutex     sync.Mutex
	online    bool
}

// getErrorMessage converts an AppError to a string
func getErrorMessage(err *model.AppError) string {
	return err.Message + " " + err.Id + " " + err.DetailedError
}

// splitTeamChannel splits a string that contains a team and a channel name
func (m *mattermost) splitTeamChannel(name string) (team, channel string) {
	tc := strings.Split(name, "/")
	switch len(tc) {
	case 1:
		// assume name only contains the channel
		channel = tc[0]
	case 2:
		// name contains team and channel
		team = tc[0]
		channel = tc[1]
	}
	return
}

// getTeam tries to identify a team by name and returns it; name can be a team
// ID or a team name; if name is empty it returns the first team the current
// user is in
func (m *mattermost) getTeam(name string) *model.Team {
	if name == "" {
		// no team name given, try to get the first team the user is in
		teams, resp := m.client.GetTeamsForUser(m.user.Id, "")
		if resp.Error != nil {
			log.Println(getErrorMessage(resp.Error))
			return nil
		}

		// return the first team if there are any teams
		if len(teams) == 0 {
			return nil
		}
		return teams[0]
	}

	// try to find team by id
	if t, resp := m.client.GetTeam(name, ""); resp.Error == nil {
		return t
	}

	// try to find team by name
	t, resp := m.client.GetTeamByName(name, "")
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return nil
	}
	return t
}

// getChannel tries to identify a channel with teamid by name and returns it;
// name can be a channel ID or a channel name
func (m *mattermost) getChannel(teamID, name string) *model.Channel {
	// try to find channel by id
	if c, resp := m.client.GetChannel(name, ""); resp.Error == nil {
		return c
	}

	// try to find channel by name
	c, resp := m.client.GetChannelByName(name, teamID, "")
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return nil
	}
	return c
}

// getUser tries to identify a user by name and returns it; name can be a user
// ID, email address or a username
func (m *mattermost) getUser(name string) *model.User {
	// try to find user by id
	if u, resp := m.client.GetUser(name, ""); resp.Error == nil {
		return u
	}

	// try to find user by email
	if u, resp := m.client.GetUserByEmail(name, ""); resp.Error == nil {
		return u
	}

	// try to find user by name
	u, resp := m.client.GetUserByUsername(name, "")
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return nil
	}
	return u
}

// createChannel creates a channel with name in team
func (m *mattermost) createChannel(team *model.Team, name string) {
	// create channel
	c := &model.Channel{
		DisplayName: name,
		Name:        name,
		Type:        model.CHANNEL_PRIVATE,
		TeamId:      team.Id,
	}
	c, resp := m.client.CreateChannel(c)
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
	}
}

// joinChannel joins channel identified by "<team>/<channel>" string
// in teamChannel
func (m *mattermost) joinChannel(teamChannel string) {
	if !m.isOnline() {
		return
	}

	// get team and channel name
	team, channel := m.splitTeamChannel(teamChannel)

	// get team id
	t := m.getTeam(team)
	if t == nil {
		log.Println("could not get team:", team, teamChannel)
		return
	}

	// check if channel already exists
	c := m.getChannel(t.Id, channel)
	if c == nil {
		// channel does not seem to exist, try to create it
		m.createChannel(t, channel)
		return
	}

	// channel exist, add current user to channel
	_, resp := m.client.AddChannelMember(c.Id, m.user.Id)
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return
	}
}

// partChannel leaves channel
func (m *mattermost) partChannel(teamChannel string) {
	if !m.isOnline() {
		return
	}

	// get team and channel name
	team, channel := m.splitTeamChannel(teamChannel)

	// get team id
	t := m.getTeam(team)
	if t == nil {
		log.Println("could not get team:", team, teamChannel)
		return
	}

	// check if channel exists
	c := m.getChannel(t.Id, channel)
	if c == nil {
		log.Println("could not get channel:", channel, teamChannel)
		return
	}

	// remove current user from channel
	_, resp := m.client.RemoveUserFromChannel(c.Id, m.user.Id)
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return
	}
}

// addChannel adds user to channel
func (m *mattermost) addChannel(teamChannel, user string) {
	if !m.isOnline() {
		return
	}

	// get team and channel name
	team, channel := m.splitTeamChannel(teamChannel)

	// get team id
	t := m.getTeam(team)
	if t == nil {
		log.Println("could not get team:", team, teamChannel)
		return
	}

	// check if channel exists
	c := m.getChannel(t.Id, channel)
	if c == nil {
		log.Println("could not get channel:", channel, teamChannel)
		return
	}

	// get user id
	u := m.getUser(user)
	if u == nil {
		log.Println("could not get user:", user)
		return
	}

	// add user to channel
	_, resp := m.client.AddChannelMember(c.Id, u.Id)
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return
	}
}

// getStatus returns our status
func (m *mattermost) getStatus() string {
	if !m.isOnline() {
		return "offline"
	}

	status, resp := m.client.GetUserStatus(m.user.Id, "")
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return ""
	}
	return status.Status
}

// setStatus sets our status
func (m *mattermost) setStatus(status string) {
	if !m.isOnline() {
		return
	}

	// check if status is valid:
	// valid user status can be online, away, offline and dnd
	switch status {
	case "online":
	case "away":
	case "offline":
	case "dnd":
	default:
		return // invalid status
	}

	// set status
	s := model.Status{
		UserId: m.user.Id,
		Status: status,
	}
	_, resp := m.client.UpdateUserStatus(m.user.Id, &s)
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
	}
}

// getChannelUsers returns a list of users in channel
func (m *mattermost) getChannelUsers(channel string) []*buddy {
	var buddies []*buddy

	if !m.isOnline() {
		return buddies
	}

	// retrieve channel members
	members, resp := m.client.GetChannelMembers(channel, 0, 60, "")
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return nil
	}

	// try to get user information of channel members
	for _, member := range *members {
		// user name
		user, resp := m.client.GetUser(member.UserId, "")
		if resp.Error != nil {
			log.Println(getErrorMessage(resp.Error))
			return nil
		}
		// user status
		status, resp := m.client.GetUserStatus(user.Id, "")
		if resp.Error != nil {
			log.Println(getErrorMessage(resp.Error))
			return nil
		}

		// add user to list
		b := newBuddy(user.Id, user.Username, status.Status)
		buddies = append(buddies, b)
	}

	return buddies
}

// getChannelName returns the name of the channel c
func (m *mattermost) getChannelName(c *model.Channel) string {
	// direct channels do not seem to set a display name; construct a name
	// from the other user's username
	if c.Type == model.CHANNEL_DIRECT {
		// get and use name of other user
		other := c.GetOtherUserIdForDM(m.user.Id)
		if other == "" {
			// there is no other user, seems like we are chatting
			// with ourselves
			return m.user.Username
		}
		user, resp := m.client.GetUser(other, "")
		if resp.Error != nil {
			// cannot retrieve username, fallback to id
			log.Println(getErrorMessage(resp.Error))
			return other
		}
		return user.Username
	}

	// use display name for other channel types
	return c.DisplayName
}

// getBuddies returns a list of teams and channels the user is in
func (m *mattermost) getBuddies() []*buddy {
	var buddies []*buddy

	if !m.isOnline() {
		return buddies
	}

	// get teams
	teams, resp := m.client.GetTeamsForUser(m.user.Id, "")
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return nil
	}
	for _, t := range teams {
		// get channels
		channels, resp := m.client.GetChannelsForTeamForUser(t.Id,
			m.user.Id, false, "")
		if resp.Error != nil {
			log.Println(getErrorMessage(resp.Error))
			return nil
		}
		for _, c := range channels {
			user := c.Id
			name := m.getChannelName(c) +
				" (" + t.DisplayName + ")"
			status := "GROUP_CHAT"
			b := newBuddy(user, name, status)
			buddies = append(buddies, b)
		}
	}

	return buddies
}

// sendMsg sends a message to channel
func (m *mattermost) sendMsg(channel string, msg string) {
	if !m.isOnline() {
		return
	}

	post := &model.Post{
		ChannelId: channel,
		Message:   msg,
	}

	if _, resp := m.client.CreatePost(post); resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
	}
}

// isOnline checks if the mattermost client is online
func (m *mattermost) isOnline() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.online
}

// setOnline sets the online state of the mattermost client
func (m *mattermost) setOnline(online bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.online = online
}

// getPostFiles returns the files attached to post as a string
func (m *mattermost) getPostFiles(post *model.Post) string {
	// return empty string if there are no files attached
	if len(post.Metadata.Files) == 0 {
		return ""
	}

	// construct and return file info string
	fileInfo := "---- Attachments:"
	for _, f := range post.Metadata.Files {
		// create link for the file
		link := m.client.ApiUrl + m.client.GetFileRoute(f.Id)

		// attach link to file name if present
		fileInfo += fmt.Sprintf(
			"\n* Name: %s\n  Type: %s\n  Size: %dB\n  Link: %s",
			f.Name, f.MimeType, f.Size, link)
	}
	return fileInfo
}

// handleWebSocketEvent handles events from the websocket
func (m *mattermost) handleWebSocketEvent(event *model.WebSocketEvent) {
	// check if event is valid
	if event == nil {
		return
	}
	log.Println("WebSocket Event:", event.EventType())

	// only handle posted events
	if event.EventType() != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	// handle post
	post := model.PostFromJson(strings.NewReader(
		event.GetData()["post"].(string)))
	if post != nil {
		// filter own messages
		if post.UserId == m.user.Id && filterOwn {
			return
		}

		// construct message text including attached files
		text := post.Message
		if fileInfo := m.getPostFiles(post); fileInfo != "" {
			if text != "" {
				text += "\n\n"
			}
			text += fileInfo
		}
		log.Println("Message:", post.CreateAt, post.ChannelId,
			post.UserId, text)

		// get name of user who sent this message
		username := post.UserId
		user, resp := m.client.GetUser(post.UserId, "")
		if resp.Error != nil {
			log.Println(getErrorMessage(resp.Error))
		} else {
			username = user.Username
		}

		// construct message with format:
		// chat: msg: <acc_id> <chat> <timestamp> <sender> <message>
		// and send it via the client queue
		msg := fmt.Sprintf("chat: msg: %d %s %d %s %s\r\n",
			m.accountID, post.ChannelId, post.CreateAt/1000,
			username, html.EscapeString(text))
		clientQueue.send(msg)
	}
}

// connect connects to a mattermost server
func (m *mattermost) connect() bool {
	// login
	log.Println("Connecting to mattermost server", m.server)
	user, resp := m.client.Login(m.username, m.password)
	if resp.Error != nil {
		log.Println(getErrorMessage(resp.Error))
		return false
	}
	log.Println("Logged in as user", user.Username)
	m.user = user

	// create websocket and start listening for events
	websock, err := model.NewWebSocketClient4(webSocketPrefix+m.server,
		m.client.AuthToken)
	if err != nil {
		log.Println(getErrorMessage(err))
		return false
	}
	m.websock = websock
	m.websock.Listen()
	m.setOnline(true)
	return true
}

// loop runs the main loop of the mattermost client handling websocket events
func (m *mattermost) loop() {
	defer m.websock.Close()

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
	for !m.connect() {
		select {
		case <-time.After(15 * time.Second):
			// wait before reconnecting
		case <-m.done:
			return
		}
	}
	m.loop()
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
		client:    model.NewAPIv4Client(httpPrefix + server),
		user:      &model.User{},
		done:      make(chan bool, 1),
	}
	return &m
}

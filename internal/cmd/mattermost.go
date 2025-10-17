package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

// teamChannel contains a team's channel and its name
type teamChannel struct {
	channel *model.Channel
	name    string
}

// teamChannels stores a mapping from team to a list of channels of the team
type teamChannels map[*model.Team][]*teamChannel

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
	history   []string
	noHistory bool

	// filterOwn toggles filtering of own messages
	filterOwn bool

	// httpPrefix is prepended to the server to form a http url
	httpPrefix string

	// webSocketPrefix is prepended to the server to form a websocket url
	webSocketPrefix string

	// teamChannels stores joined channels for each team
	teamChannels teamChannels

	// channels stores information of joined channels
	channels *channels
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

// getTeamByID tries to get a team by its ID
func (m *mattermost) getTeamByID(ctx context.Context, id string) *model.Team {
	if !model.IsValidId(id) {
		return nil
	}
	t, _, err := m.client.GetTeam(ctx, id, "")
	if err != nil {
		return nil
	}
	return t
}

// getTeamByName tries to get a team by its name
func (m *mattermost) getTeamByName(ctx context.Context, name string) *model.Team {
	if !model.IsValidTeamName(name) {
		return nil
	}
	t, _, err := m.client.GetTeamByName(ctx, name, "")
	if err != nil {
		return nil
	}
	return t
}

// getTeam tries to identify a team by name and returns it; name can be a team
// ID or a team name; if name is empty it returns the first team the current
// user is in
func (m *mattermost) getTeam(ctx context.Context, name string) *model.Team {
	if name == "" {
		// no team name given, try to get the first team the user is in
		teams := []*model.Team{}
		for t := range m.getTeamChannels() {
			teams = append(teams, t)
		}
		if len(teams) == 0 {
			return nil
		}
		return teams[0]
	}

	// try to find team by id
	if t := m.getTeamByID(ctx, name); t != nil {
		return t
	}

	// try to find team by name
	return m.getTeamByName(ctx, name)
}

// getChannelByID tries to get a channel by its ID
func (m *mattermost) getChannelByID(ctx context.Context, id string) *model.Channel {
	if !model.IsValidId(id) {
		return nil
	}
	c, _, err := m.client.GetChannel(ctx, id, "")
	if err != nil {
		return nil
	}
	return c
}

// getChannelByName tries to get a channel by its team ID and name
func (m *mattermost) getChannelByName(ctx context.Context, teamID, name string) *model.Channel {
	if !model.IsValidChannelIdentifier(name) {
		return nil
	}
	c, _, err := m.client.GetChannelByName(ctx, name, teamID, "")
	if err != nil {
		return nil
	}
	return c
}

// getChannel tries to identify a channel with teamid by name and returns it;
// name can be a channel ID or a channel name
func (m *mattermost) getChannel(ctx context.Context, teamID, name string) *model.Channel {
	// try to find channel by id
	if c := m.getChannelByID(ctx, name); c != nil {
		return c
	}

	// try to find channel by name
	return m.getChannelByName(ctx, teamID, name)
}

// getUserByID tries to get a user by its ID
func (m *mattermost) getUserByID(ctx context.Context, id string) *model.User {
	if !model.IsValidId(id) {
		return nil
	}
	u, _, err := m.client.GetUser(ctx, id, "")
	if err != nil {
		return nil
	}
	return u
}

// getUserByEmail tries to get a user by its email address
func (m *mattermost) getUserByEmail(ctx context.Context, email string) *model.User {
	if !model.IsValidEmail(email) {
		return nil
	}
	u, _, err := m.client.GetUserByEmail(ctx, email, "")
	if err != nil {
		return nil
	}
	return u
}

// getUserByUsername tries to get a user by its username
func (m *mattermost) getUserByUsername(ctx context.Context, username string) *model.User {
	if !model.IsValidUsername(username) {
		return nil
	}
	u, _, err := m.client.GetUserByUsername(ctx, username, "")
	if err != nil {
		return nil
	}
	return u
}

// getUser tries to identify a user by name and returns it; name can be a user
// ID, email address or a username
func (m *mattermost) getUser(ctx context.Context, name string) *model.User {
	// try to find user by id
	if u := m.getUserByID(ctx, name); u != nil {
		return u
	}

	// try to find user by email
	if u := m.getUserByEmail(ctx, name); u != nil {
		return u
	}

	// try to find user by name
	return m.getUserByUsername(ctx, name)
}

// createChannel creates a channel with name in team
func (m *mattermost) createChannel(ctx context.Context, team *model.Team, name string) {
	// create channel
	c := &model.Channel{
		DisplayName: name,
		Name:        name,
		Type:        model.ChannelTypePrivate,
		TeamId:      team.Id,
	}
	_, _, err := m.client.CreateChannel(ctx, c)
	if err != nil {
		logError(err)
	}
}

// joinChannel joins channel identified by "<team>/<channel>" string
// in teamChannel
func (m *mattermost) joinChannel(ctx context.Context, teamChannel string) {
	if !m.isOnline() {
		return
	}

	// get team and channel name
	team, channel := m.splitTeamChannel(teamChannel)

	// get team id
	t := m.getTeam(ctx, team)
	if t == nil {
		logError("could not get team:", team, teamChannel)
		return
	}

	// check if channel already exists
	c := m.getChannel(ctx, t.Id, channel)
	if c == nil {
		// channel does not seem to exist, try to create it
		m.createChannel(ctx, t, channel)
		return
	}

	// channel exist, add current user to channel
	_, _, err := m.client.AddChannelMember(ctx, c.Id, m.user.Id)
	if err != nil {
		logError(err)
		return
	}
}

// partChannel leaves channel
func (m *mattermost) partChannel(ctx context.Context, teamChannel string) {
	if !m.isOnline() {
		return
	}

	// get team and channel name
	team, channel := m.splitTeamChannel(teamChannel)

	// get team id
	t := m.getTeam(ctx, team)
	if t == nil {
		logError("could not get team:", team, teamChannel)
		return
	}

	// check if channel exists
	c := m.getChannel(ctx, t.Id, channel)
	if c == nil {
		logError("could not get channel:", channel, teamChannel)
		return
	}

	// remove current user from channel
	_, err := m.client.RemoveUserFromChannel(ctx, c.Id, m.user.Id)
	if err != nil {
		logError(err)
		return
	}
}

// addChannel adds user to channel
func (m *mattermost) addChannel(ctx context.Context, teamChannel, user string) {
	if !m.isOnline() {
		return
	}

	// get team and channel name
	team, channel := m.splitTeamChannel(teamChannel)

	// get team id
	t := m.getTeam(ctx, team)
	if t == nil {
		logError("could not get team:", team, teamChannel)
		return
	}

	// check if channel exists
	c := m.getChannel(ctx, t.Id, channel)
	if c == nil {
		logError("could not get channel:", channel, teamChannel)
		return
	}

	// get user id
	u := m.getUser(ctx, user)
	if u == nil {
		logError("could not get user:", user)
		return
	}

	// add user to channel
	_, _, err := m.client.AddChannelMember(ctx, c.Id, u.Id)
	if err != nil {
		logError(err)
		return
	}
}

// getStatus returns our status
func (m *mattermost) getStatus(ctx context.Context) string {
	if !m.isOnline() {
		return "offline"
	}

	status, _, err := m.client.GetUserStatus(ctx, m.user.Id, "")
	if err != nil {
		logError(err)
		return ""
	}
	return status.Status
}

// setStatus sets our status
func (m *mattermost) setStatus(ctx context.Context, status string) {
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
	_, _, err := m.client.UpdateUserStatus(ctx, m.user.Id, &s)
	if err != nil {
		logError(err)
	}
}

// getChannelUsers returns a list of users in channel
func (m *mattermost) getChannelUsers(ctx context.Context, channel string) []*buddy {
	var buddies []*buddy

	if !m.isOnline() {
		return buddies
	}

	// retrieve channel members
	members, _, err := m.client.GetChannelMembers(ctx, channel, 0, 60, "")
	if err != nil {
		logError(err)
		return nil
	}

	// try to get user information of channel members
	for _, member := range members {
		// user name
		user, _, err := m.client.GetUser(ctx, member.UserId, "")
		if err != nil {
			logError(err)
			return nil
		}
		// user status
		status, _, err := m.client.GetUserStatus(ctx, user.Id, "")
		if err != nil {
			logError(err)
			return nil
		}

		// add user to list
		b := newBuddy(user.Id, user.Username, status.Status)
		buddies = append(buddies, b)
	}

	return buddies
}

// getChannelName returns the name of the channel c
func (m *mattermost) getChannelName(ctx context.Context, c *model.Channel) string {
	// direct channels do not seem to set a display name; construct a name
	// from the other user's username
	if c.Type == model.ChannelTypeDirect {
		// get and use name of other user
		other := c.GetOtherUserIdForDM(m.user.Id)
		if other == "" {
			// there is no other user, seems like we are chatting
			// with ourselves
			return m.user.Username
		}
		user, _, err := m.client.GetUser(ctx, other, "")
		if err != nil {
			// cannot retrieve username, fallback to id
			logError(err)
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

	for _, teamChannels := range m.getTeamChannels() {
		// get channels
		for _, tc := range teamChannels {
			user := tc.channel.Id
			name := tc.name
			status := "GROUP_CHAT"
			b := newBuddy(user, name, status)
			buddies = append(buddies, b)
		}
	}

	return buddies
}

// sendMsg sends a message to channel
func (m *mattermost) sendMsg(ctx context.Context, channel string, msg string) {
	if !m.isOnline() {
		return
	}

	post := &model.Post{
		ChannelId: channel,
		Message:   msg,
	}

	if _, _, err := m.client.CreatePost(ctx, post); err != nil {
		logError(err)
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

// setTeamChannels sets the map of teams and their channels
func (m *mattermost) setTeamChannels(t teamChannels) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.teamChannels = t
}

// getTeamChannels gets the map of teams and their channels
func (m *mattermost) getTeamChannels() teamChannels {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.teamChannels
}

// updateTeamChannels updates the teams and their channels
func (m *mattermost) updateTeamChannels(ctx context.Context) bool {
	// get teams
	teams, _, err := m.client.GetTeamsForUser(ctx, m.user.Id, "")
	if err != nil {
		logError(err)
		return false
	}

	// get channels
	teamChannels := make(teamChannels)
	for _, t := range teams {
		// get channels
		channels, _, err := m.client.GetChannelsForTeamForUser(
			ctx, t.Id, m.user.Id, false, "")
		if err != nil {
			logError(err)
			return false
		}

		for _, c := range channels {
			// get name of the channel
			name := m.getChannelName(ctx, c) +
				" (" + t.DisplayName + ")"
			tc := &teamChannel{c, name}

			teamChannels[t] = append(teamChannels[t], tc)
		}
	}

	// update teams and channels
	m.setTeamChannels(teamChannels)

	return true
}

// addHistory adds msg to the account history
func (m *mattermost) addHistory(msg string) {
	if m.noHistory {
		return
	}
	// only add "chat: msg:" or "message:" messages
	if !strings.HasPrefix(msg, "chat: msg:") &&
		!strings.HasPrefix(msg, "message:") {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.history = append(m.history, msg)
}

// getHistory retrieves the account histoy
func (m *mattermost) getHistory() {
	if m.noHistory {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// send all messages to client
	for _, msg := range m.history {
		clientQueue.send(msg)
	}
}

// getPostFiles returns the files attached to post as a string
func (m *mattermost) getPostFiles(ctx context.Context, post *model.Post) string {
	// return empty string if there are no files attached
	if len(post.Metadata.Files) == 0 {
		return ""
	}

	// construct and return file info string
	fileInfo := "---- Attachments:"
	for _, f := range post.Metadata.Files {
		// create link for the file
		link, _, err := m.client.GetFileLink(ctx, f.Id)
		if err != nil {
			logError(err)
			continue
		}

		// attach link to file name if present
		fileInfo += fmt.Sprintf(
			"\n* Name: %s\n  Type: %s\n  Size: %dB\n  Link: %s",
			f.Name, f.MimeType, f.Size, link)
	}
	return fileInfo
}

// handlePost handles the post
func (m *mattermost) handlePost(ctx context.Context, post *model.Post) {
	// filter own messages
	if post.UserId == m.user.Id && m.filterOwn {
		return
	}

	// construct message text including attached files
	text := post.Message
	if fileInfo := m.getPostFiles(ctx, post); fileInfo != "" {
		if text != "" {
			text += "\n\n"
		}
		text += fileInfo
	}
	logDebug("Message:", post.CreateAt, post.ChannelId,
		post.UserId, text)

	// get name of user who sent this message
	username := post.UserId
	if post.UserId == m.user.Id {
		username = "<self>"
	} else {
		user, _, err := m.client.GetUser(ctx, post.UserId, "")
		if err != nil {
			logError(err)
		} else {
			username = user.Username
		}
	}

	// construct message with format:
	// chat: msg: <acc_id> <chat> <timestamp> <sender> <message>
	// and send it via the client queue
	msg := fmt.Sprintf("chat: msg: %d %s %d %s %s\r\n",
		m.accountID, post.ChannelId, post.CreateAt/1000,
		username, html.EscapeString(text))
	m.addHistory(msg)
	clientQueue.send(msg)

	// save last post id of channel
	m.channels.updatePostID(post.ChannelId, post.Id)
}

// handleRemoved handles user removed events
func (m *mattermost) handleRemoved(event *model.WebSocketEvent) {
	data := event.GetData()
	userID := data["user_id"]
	if userID != nil && userID != m.user.Id {
		return
	}
	chanID, ok := data["channel_id"].(string)
	if !ok {
		return
	}

	// we are removed from channel, remove stored channel
	// information
	m.channels.deleteChannel(chanID)
}

// handleTeamChannelChange handles team and channel change events
func (m *mattermost) handleTeamChannelChange(ctx context.Context, event *model.WebSocketEvent) {
	// handle removed events
	if event.EventType() == model.WebsocketEventUserRemoved {
		m.handleRemoved(event)
	}

	// update teams and channels
	m.updateTeamChannels(ctx)
}

// handleWebSocketEvent handles events from the websocket
func (m *mattermost) handleWebSocketEvent(ctx context.Context, event *model.WebSocketEvent) {
	// check if event is valid
	if event == nil {
		return
	}
	logDebug("WebSocket Event:", event.EventType())

	// handle special events
	switch event.EventType() {

	// handle team change events
	case model.WebsocketEventAddedToTeam,
		model.WebsocketEventLeaveTeam,
		model.WebsocketEventUpdateTeam,
		model.WebsocketEventDeleteTeam,
		model.WebsocketEventRestoreTeam:
		m.handleTeamChannelChange(ctx, event)
		return

	// handle channel change events
	case model.WebsocketEventChannelConverted,
		model.WebsocketEventChannelCreated,
		model.WebsocketEventChannelDeleted,
		model.WebsocketEventChannelUpdated,
		model.WebsocketEventChannelMemberUpdated:
		m.handleTeamChannelChange(ctx, event)
		return

	// hande user change events
	case model.WebsocketEventUserAdded,
		model.WebsocketEventUserUpdated,
		model.WebsocketEventUserRoleUpdated,
		model.WebsocketEventMemberroleUpdated,
		model.WebsocketEventUserRemoved:
		m.handleTeamChannelChange(ctx, event)
		return
	}

	// only handle posted events from this point on
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	// handle post
	var post *model.Post
	data := strings.NewReader(event.GetData()["post"].(string))
	if err := json.NewDecoder(data).Decode(&post); err != nil {
		logError(err)
		return
	}
	if post != nil {
		m.handlePost(ctx, post)
	}
}

// getOldChannelMessages retrieves old/unread messages of the channel
// identified by id
func (m *mattermost) getOldChannelMessages(ctx context.Context, id string) {
	// get last known post id of channel
	postID := m.channels.getPostID(id)
	for {
		// get batch of message after last know post id
		posts, _, err := m.client.GetPostsAfter(ctx,
			id, postID, 0, 60, "", false, true)
		if err != nil {
			logError(err)
			return
		}

		// reverse message order
		for i := len(posts.Order) - 1; i >= 0; i-- {
			p := posts.Order[i]
			m.handlePost(ctx, posts.Posts[p])
			postID = p

		}
		if posts.NextPostId == "" {
			break
		}
	}
}

// getOldMessages retrieves old/unread messages
func (m *mattermost) getOldMessages(ctx context.Context) {
	for _, teamChannels := range m.getTeamChannels() {
		// get messages in each channel
		for _, tc := range teamChannels {
			m.getOldChannelMessages(ctx, tc.channel.Id)
		}
	}
}

// connect connects to a mattermost server
func (m *mattermost) connect(ctx context.Context) bool {
	// login
	logInfo("Connecting to mattermost server", m.server)
	ctxLogin, cancelLogin := context.WithTimeout(ctx, 30*time.Second)
	defer cancelLogin()
	user, _, err := m.client.Login(ctxLogin, m.username, m.password)
	if err != nil {
		logError(err)
		return false
	}
	logInfo("Logged in as user", user.Username)
	m.user = user

	// update teams and channels
	ctxTeams, cancelTeams := context.WithTimeout(ctx, time.Minute)
	defer cancelTeams()
	if !m.updateTeamChannels(ctxTeams) {
		return false
	}

	// retrieve unread messages
	ctxMsgs, cancelMsgs := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelMsgs()
	m.getOldMessages(ctxMsgs)

	// create websocket and start listening for events
	websock, err := model.NewWebSocketClient4(m.webSocketPrefix+m.server,
		m.client.AuthToken)
	if err != nil {
		logError(err)
		return false
	}
	m.websock = websock
	m.websock.Listen()
	m.setOnline(true)
	return true
}

// loop runs the main loop of the mattermost client handling websocket events
func (m *mattermost) loop(ctx context.Context) bool {
	defer m.websock.Close()

	// handle websocket events
	for {
		select {
		case event, more := <-m.websock.EventChannel:
			if !more {
				// event channel was closed unexpectedly,
				// log error if present, set client offline
				// and return an error to trigger a reconnect
				if err := m.websock.ListenError; err != nil {
					logError(getErrorMessage(err))
				}
				m.setOnline(false)
				return false
			}

			// handle event
			m.handleWebSocketEvent(ctx, event)
		case <-m.websock.PingTimeoutChannel:
			logError("websocket ping timeout")
		case <-m.done:
			return true
		}
	}
}

// run starts the mattermost client
func (m *mattermost) run(ctx context.Context) {
	for {
		// try to (re)connect to the server
		for !m.connect(ctx) {
			select {
			case <-time.After(15 * time.Second):
				// wait before reconnecting
			case <-m.done:
				return
			}
		}

		// connection established, run main loop until we are done;
		// if there is an error, reconnect to the server
		if m.loop(ctx) {
			return
		}
	}
}

// stop shuts down the mattermost client
func (m *mattermost) stop() {
	m.done <- true
}

// newClient creates a new mattermost client
func newClient(config *Config, accountID int, server, username,
	password string) *mattermost {

	// configure encryption
	httpPrefix := "https://"
	webSocketPrefix := "wss://"
	if config.DisableEncryption {
		httpPrefix = "http://"
		webSocketPrefix = "ws://"
	}

	m := mattermost{
		accountID: accountID,
		server:    server,
		username:  username,
		password:  password,
		client:    model.NewAPIv4Client(httpPrefix + server),
		done:      make(chan bool, 1),

		filterOwn:       config.FilterOwn,
		httpPrefix:      httpPrefix,
		webSocketPrefix: webSocketPrefix,
		noHistory:       config.DisableHistory,
		channels:        newChannels(accountID),
	}
	return &m
}

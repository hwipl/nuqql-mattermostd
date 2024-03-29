package cmd

import (
	"bufio"
	"context"
	"fmt"
	"html"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	helpMessage = `info: List of commands and their description:
account list
    list all accounts and their account ids.
account add <protocol> <user> <password>
    add a new account for chat protocol <protocol> with user name <user> and
    the password <password>. The supported chat protocol(s) are backend
    specific. The user name is chat protocol specific. An account id is
    assigned to the account that can be shown with "account list".
account <id> delete
    delete the account with the account id <id>.
account <id> buddies [online]
    list all buddies on the account with the account id <id>. Optionally, show
    only online buddies with the extra parameter "online".
account <id> collect
    collect all messages received on the account with the account id <id>.
account <id> send <user> <msg>
    send a message to the user <user> on the account with the account id <id>.
account <id> status get
    get the status of the account with the account id <id>.
account <id> status set <status>
    set the status of the account with the account id <id> to <status>.
account <id> chat list
    list all group chats on the account with the account id <id>.
account <id> chat join <chat>
    join the group chat <chat> on the account with the account id <id>.
account <id> chat part <chat>
    leave the group chat <chat> on the account with the account id <id>.
account <id> chat send <chat> <msg>
    send the message <msg> to the group chat <chat> on the account with the
    account id <id>.
account <id> chat users <chat>
    list the users in the group chat <chat> on the account with the
    account id <id>.
account <id> chat invite <chat> <user>
    invite the user <user> to the group chat <chat> on the account with the
    account id <id>.
version
    get version of the backend
bye
    disconnect from backend
quit
    quit backend
help
    show this help` + "\r\n"
)

var (
	// brRegex is a regular expression for <br/> tags
	brRegex = regexp.MustCompile("(?i)<br/>")
)

// server stores server information
type server struct {
	network  string
	address  string
	listener net.Listener
	conn     net.Conn

	// is server/client active?
	serverActive bool
	clientActive bool
}

// sendClient sends msg to the client
func (s *server) sendClient(msg string) {
	clientQueue.send(msg)
}

// createAccountMessage creates an account message for account a
func createAccountMessage(a *account) string {
	// get account status
	status := "offline"
	if a.client != nil && a.client.isOnline() {
		status = "online"
	}

	// return message with the following format:
	// account: <id> <name> <protocol> <user> <status>
	return fmt.Sprintf("account: %d %s %s %s %s\r\n", a.ID, "()",
		a.Protocol, a.User, status)
}

// getAccountListMessages returns the account list as a string of messages
func (s *server) getAccountListMessages() (messages string) {
	accounts := getAccounts()
	for _, a := range accounts {
		messages += createAccountMessage(a)
	}
	messages += "info: listed accounts.\r\n"
	if len(accounts) == 0 {
		h := "info: You do not have any accounts configured.\r\n"
		h += "info: You can add a new mattermost account with the " +
			"following command: " +
			"account add mattermost <username>@<server> " +
			"<password>\r\n"
		h += "info: Example: account add mattermost " +
			"dummy@yourserver.org:8065 YourPassword\r\n"
		messages += h
	}
	return
}

// handleAccountList handles an account list command
func (s *server) handleAccountList() {
	// send messages as replies
	r := s.getAccountListMessages()
	logDebug(r)
	s.sendClient(r)
}

// handleAccountAdd handles an account add command
func (s *server) handleAccountAdd(ctx context.Context, parts []string) {
	// expected command format:
	// account add <protocol> <user> <password>
	if len(parts) < 5 {
		return
	}
	protocol := parts[2]
	user := parts[3]
	password := parts[4]
	id := addAccount(ctx, protocol, user, password)
	logInfo("added new account with id:", id)

	// optional reply:
	// info: new account added.
	s.sendClient(fmt.Sprintf("info: added account %d.\r\n", id))
	if conf.PushAccounts {
		// send account message with push accounts enabled
		a := getAccount(id)
		m := createAccountMessage(a)
		s.sendClient(m)
	}
}

// handleAccountDelete handles an account delete command
func (s *server) handleAccountDelete(id int) {
	if delAccount(id) {
		logInfo("deleted account with id: ", id)
		s.sendClient(fmt.Sprintf("info: account %d deleted.\r\n", id))
	}

}

// handleAccountBuddies handles an account buddies command
func (s *server) handleAccountBuddies(a *account) {
	for _, b := range a.client.getBuddies() {
		//buddy: <acc_id> status: <status> name: <name> alias: [alias]
		m := fmt.Sprintf("buddy: %d status: %s name: %s alias: %s\r\n",
			a.ID, b.status, b.user, url.PathEscape(b.name))
		s.sendClient(m)
	}
}

// handleAccountCollect handles an account collect command
func (s *server) handleAccountCollect(a *account) {
	a.client.getHistory()
}

// unescapeMessage converts nuqql message to original format:
// nuqql sends html-escaped messages with newlines replaced by <br/>
func unescapeMessage(msg string) string {
	msg = brRegex.ReplaceAllString(msg, "\n")
	msg = html.UnescapeString(msg)
	return msg
}

// handleAccountSend handles an account send command
func (s *server) handleAccountSend(ctx context.Context, a *account, parts []string) {
	// account <id> send <user> <msg>
	if len(parts) < 5 {
		return
	}
	channel := parts[3]
	msg := strings.Join(parts[4:], " ")
	logDebug("sending message to channel "+channel+":", msg)
	a.client.sendMsg(ctx, channel, unescapeMessage(msg))
}

// handleAccountStatusGet handles an account status get command
func (s *server) handleAccountStatusGet(ctx context.Context, a *account) {
	// account <id> status get
	status := a.client.getStatus(ctx)
	if status == "" {
		return
	}

	// create and send status message with format:
	// status: account <acc_id> status: <status>
	m := fmt.Sprintf("status: account %d status: %s\r\n", a.ID, status)
	s.sendClient(m)
}

// handleAccountStatusSet handles an account status set command
func (s *server) handleAccountStatusSet(ctx context.Context, a *account, parts []string) {
	// account <id> status set <status>
	if len(parts) < 5 {
		return
	}

	// try to set status
	status := parts[4]
	a.client.setStatus(ctx, status)
}

// handleAccountStatus handles an account status command
func (s *server) handleAccountStatus(ctx context.Context, a *account, parts []string) {
	// status commands have at least 4 parts
	if len(parts) < 4 {
		return
	}

	// handle status commands
	switch parts[3] {
	case "get":
		s.handleAccountStatusGet(ctx, a)
	case "set":
		s.handleAccountStatusSet(ctx, a, parts)
	}
}

// handleAccountChatList handles an account chat list command
func (s *server) handleAccountChatList(a *account) {
	for _, b := range a.client.getBuddies() {
		// chat: list: <acc_id> <chat_id> <chat_alias> <nick>
		m := fmt.Sprintf("chat: list: %d %s %s %s\r\n",
			a.ID, b.user, url.PathEscape(b.name),
			a.client.username)
		s.sendClient(m)
	}
}

// handleAccountChatJoin handles an account chat join command
func (s *server) handleAccountChatJoin(ctx context.Context, a *account, parts []string) {
	// account <id> chat join <chat>
	if len(parts) < 5 {
		return
	}
	channel := parts[4]
	logInfo("joining channel " + channel)
	a.client.joinChannel(ctx, channel)
}

// handleAccountChatPart handles an account chat part command
func (s *server) handleAccountChatPart(ctx context.Context, a *account, parts []string) {
	// account <id> chat part <chat>
	if len(parts) < 5 {
		return
	}
	channel := parts[4]
	logInfo("leaving channel " + channel)
	a.client.partChannel(ctx, channel)
}

// handleAccountChatSend handles an account chat send command
func (s *server) handleAccountChatSend(ctx context.Context, a *account, parts []string) {
	// account <id> chat send <chat> <msg>
	if len(parts) < 6 {
		return
	}
	channel := parts[4]
	msg := strings.Join(parts[5:], " ")
	logDebug("sending message to channel "+channel+":", msg)
	a.client.sendMsg(ctx, channel, unescapeMessage(msg))
}

// handleAccountChat handles an account chat users command
func (s *server) handleAccountChatUsers(ctx context.Context, a *account, parts []string) {
	// account <id> chat users <chat>
	if len(parts) < 5 {
		return
	}

	channel := parts[4]
	users := a.client.getChannelUsers(ctx, channel)
	for _, u := range users {
		// create and send message with format:
		// chat: user: <acc_id> <chat> <name> <alias> <state>
		m := fmt.Sprintf("chat: user: %d %s %s %s %s\r\n",
			a.ID, channel, u.user, url.PathEscape(u.name),
			u.status)
		s.sendClient(m)
	}
}

// handleAccountChatInvite handles an account chat invite command
func (s *server) handleAccountChatInvite(ctx context.Context, a *account, parts []string) {
	// account <id> chat invite <chat> <user>
	if len(parts) < 6 {
		return
	}

	channel := parts[4]
	user := parts[5]
	logInfo("adding " + user + " to channel " + channel)
	a.client.addChannel(ctx, channel, user)
}

// handleAccountChat handles an account chat command
func (s *server) handleAccountChat(ctx context.Context, a *account, parts []string) {
	// chat commands have at least 4 parts
	if len(parts) < 4 {
		return
	}

	// handle chat subcommands
	switch parts[3] {
	case "list":
		s.handleAccountChatList(a)
	case "join":
		s.handleAccountChatJoin(ctx, a, parts)
	case "part":
		s.handleAccountChatPart(ctx, a, parts)
	case "send":
		s.handleAccountChatSend(ctx, a, parts)
	case "users":
		s.handleAccountChatUsers(ctx, a, parts)
	case "invite":
		s.handleAccountChatInvite(ctx, a, parts)
	}
}

// handleAccountCommand handles an account command received from the client
func (s *server) handleAccountCommand(ctx context.Context, parts []string) {
	// account commands consist of at least 2 parts
	if len(parts) < 2 {
		return
	}

	// commands "list" and "add" are the only ones that do not start with
	// an account id; handle them first
	if parts[1] == "list" {
		s.handleAccountList()
		return
	}
	if parts[1] == "add" {
		s.handleAccountAdd(ctx, parts)
		return
	}

	// other commands contain at least 3 parts
	if len(parts) < 3 {
		return
	}

	// other commands contain an account id; try to parse it
	id, err := strconv.ParseUint(parts[1], 10, 16)
	if err != nil {
		return
	}

	// check if there is an account with this id
	a := getAccount(int(id))
	if a == nil {
		return
	}

	// handle other commands
	switch parts[2] {
	case "delete":
		s.handleAccountDelete(a.ID)
	case "buddies":
		s.handleAccountBuddies(a)
	case "collect":
		s.handleAccountCollect(a)
	case "send":
		s.handleAccountSend(ctx, a, parts)
	case "status":
		s.handleAccountStatus(ctx, a, parts)
	case "chat":
		s.handleAccountChat(ctx, a, parts)
	}
}

// handleVersionCommand handles a version command received from the client
func (s *server) handleVersionCommand() {
	versionFmt := "info: version: %s v%s\r\n"
	msg := fmt.Sprintf(versionFmt, conf.Name, backendVersion)
	s.sendClient(msg)
}

// handleCommand handles a command received from the client
func (s *server) handleCommand(ctx context.Context, cmd string) {
	logDebug("client:", cmd)

	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "account":
		s.handleAccountCommand(ctx, parts)
	case "version":
		s.handleVersionCommand()
	case "bye":
		s.clientActive = false
	case "quit":
		s.clientActive = false
		s.serverActive = false
	case "help":
		s.sendClient(helpMessage)
	}
}

// sendEarly sends msg to client, should only be used before client queue is
// active
func (s *server) sendEarly(msg string) {
	w := bufio.NewWriter(s.conn)
	n, err := w.WriteString(msg)
	if n < len(msg) || err != nil {
		logError(err)
		return
	}
	if err := w.Flush(); err != nil {
		logError(err)
		return
	}
}

// handleClient handles a single client connection
func (s *server) handleClient() {
	defer func() {
		if err := s.conn.Close(); err != nil {
			logError(err)
		}
	}()
	logInfo("New client connection", s.conn.RemoteAddr())

	// send welcome message to client
	s.sendEarly(fmt.Sprintf("info: Welcome to nuqql-mattermostd v%s!\r\n",
		backendVersion))
	s.sendEarly("info: Enter \"help\" for a list of available commands " +
		"and their help texts\r\n")

	// if push accounts is enabled, send list of accounts to client
	if conf.PushAccounts {
		s.sendEarly("info: Listing your accounts:\r\n")
		s.sendEarly(s.getAccountListMessages())
	}

	// configure client in queue
	clientQueue.setClient(s.conn)
	defer clientQueue.setClient(nil)

	// enable client
	s.clientActive = true

	// start client command handling loop
	r := bufio.NewReader(s.conn)
	c := ""
	ctx := context.Background()
	for s.clientActive {
		// read a cmd line from the client
		cmd, err := r.ReadString('\n')
		if err != nil {
			logError("client:", err)
			return
		}

		// read and concatenate cmd lines until "\r\n"
		c += cmd
		if len(c) >= 2 && c[len(c)-2] == '\r' {
			s.handleCommand(ctx, c[:len(c)-2])
			c = ""
		}
	}
}

// run runs the server
func (s *server) run() {
	if s.network == "unix" {
		// remove old socket file
		if err := os.Remove(s.address); err != nil {
			logError(err)
		}
	}

	// start listener
	l, err := net.Listen(s.network, s.address)
	if err != nil {
		logFatal(err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			logError(err)
		}
	}()
	s.listener = l
	s.serverActive = true

	// handle client connections
	logInfo("Server waiting for client connection")
	for s.serverActive {
		conn, err := s.listener.Accept()
		if err != nil {
			logError(err)
			continue
		}
		s.conn = conn

		// only one client connection is allowed at the same time
		s.handleClient()
	}
}

// runServer runs the server that handles nuqql/telnet connections
func runServer() {
	server := server{
		network: conf.GetListenNetwork(),
		address: conf.GetListenAddress(),
	}
	server.run()
}

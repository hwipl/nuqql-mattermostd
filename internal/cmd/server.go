package cmd

import (
	"bufio"
	"fmt"
	"html"
	"net"
	"net/url"
	"os"
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

// handleAccountList handles an account list command
func (s *server) handleAccountList() {
	for _, a := range getAccounts() {
		// get account status
		status := "offline"
		if a.client != nil && a.client.isOnline() {
			status = "online"
		}

		// send replies with the following format:
		// account: <id> <name> <protocol> <user> <status>
		r := fmt.Sprintf("account: %d %s %s %s %s\r\n", a.ID, "()",
			a.Protocol, a.User, status)
		logDebug(r)
		s.sendClient(r)
	}
}

// handleAccountAdd handles an account add command
func (s *server) handleAccountAdd(parts []string) {
	// expected command format:
	// account add <protocol> <user> <password>
	if len(parts) < 5 {
		return
	}
	protocol := parts[2]
	user := parts[3]
	password := parts[4]
	id := addAccount(protocol, user, password)
	logInfo("added new account with id:", id)

	// optional reply:
	// info: new account added.
	s.sendClient("info: new account added.\r\n")
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
	clientQueue.getHistory()
}

// handleAccountSend handles an account send command
func (s *server) handleAccountSend(a *account, parts []string) {
	// account <id> send <user> <msg>
	if len(parts) < 5 {
		return
	}
	channel := parts[3]
	msg := strings.Join(parts[4:], " ")
	logDebug("sending message to channel "+channel+":", msg)
	a.client.sendMsg(channel, html.UnescapeString(msg))
}

// handleAccountStatusGet handles an account status get command
func (s *server) handleAccountStatusGet(a *account) {
	// account <id> status get
	status := a.client.getStatus()
	if status == "" {
		return
	}

	// create and send status message with format:
	// status: account <acc_id> status: <status>
	m := fmt.Sprintf("status: account %d status: %s\r\n", a.ID, status)
	s.sendClient(m)
}

// handleAccountStatusSet handles an account status set command
func (s *server) handleAccountStatusSet(a *account, parts []string) {
	// account <id> status set <status>
	if len(parts) < 5 {
		return
	}

	// try to set status
	status := parts[4]
	a.client.setStatus(status)
}

// handleAccountStatus handles an account status command
func (s *server) handleAccountStatus(a *account, parts []string) {
	// status commands have at least 4 parts
	if len(parts) < 4 {
		return
	}

	// handle status commands
	switch parts[3] {
	case "get":
		s.handleAccountStatusGet(a)
	case "set":
		s.handleAccountStatusSet(a, parts)
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
func (s *server) handleAccountChatJoin(a *account, parts []string) {
	// account <id> chat join <chat>
	if len(parts) < 5 {
		return
	}
	channel := parts[4]
	logInfo("joining channel " + channel)
	a.client.joinChannel(channel)
}

// handleAccountChatPart handles an account chat part command
func (s *server) handleAccountChatPart(a *account, parts []string) {
	// account <id> chat part <chat>
	if len(parts) < 5 {
		return
	}
	channel := parts[4]
	logInfo("leaving channel " + channel)
	a.client.partChannel(channel)
}

// handleAccountChatSend handles an account chat send command
func (s *server) handleAccountChatSend(a *account, parts []string) {
	// account <id> chat send <chat> <msg>
	if len(parts) < 6 {
		return
	}
	channel := parts[4]
	msg := strings.Join(parts[5:], " ")
	logDebug("sending message to channel "+channel+":", msg)
	a.client.sendMsg(channel, html.UnescapeString(msg))
}

// handleAccountChat handles an account chat users command
func (s *server) handleAccountChatUsers(a *account, parts []string) {
	// account <id> chat users <chat>
	if len(parts) < 5 {
		return
	}

	channel := parts[4]
	users := a.client.getChannelUsers(channel)
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
func (s *server) handleAccountChatInvite(a *account, parts []string) {
	// account <id> chat invite <chat> <user>
	if len(parts) < 6 {
		return
	}

	channel := parts[4]
	user := parts[5]
	logInfo("adding " + user + " to channel " + channel)
	a.client.addChannel(channel, user)
}

// handleAccountChat handles an account chat command
func (s *server) handleAccountChat(a *account, parts []string) {
	// chat commands have at least 4 parts
	if len(parts) < 4 {
		return
	}

	// handle chat subcommands
	switch parts[3] {
	case "list":
		s.handleAccountChatList(a)
	case "join":
		s.handleAccountChatJoin(a, parts)
	case "part":
		s.handleAccountChatPart(a, parts)
	case "send":
		s.handleAccountChatSend(a, parts)
	case "users":
		s.handleAccountChatUsers(a, parts)
	case "invite":
		s.handleAccountChatInvite(a, parts)
	}
}

// handleAccountCommand handles an account command received from the client
func (s *server) handleAccountCommand(parts []string) {
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
		s.handleAccountAdd(parts)
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
		s.handleAccountSend(a, parts)
	case "status":
		s.handleAccountStatus(a, parts)
	case "chat":
		s.handleAccountChat(a, parts)
	}
}

// handleVersionCommand handles a version command received from the client
func (s *server) handleVersionCommand() {
	versionFmt := "info: version: %s v%s\r\n"
	msg := fmt.Sprintf(versionFmt, conf.Name, backendVersion)
	s.sendClient(msg)
}

// handleCommand handles a command received from the client
func (s *server) handleCommand(cmd string) {
	logDebug("client:", cmd)

	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "account":
		s.handleAccountCommand(parts)
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

// handleClient handles a single client connection
func (s *server) handleClient() {
	defer s.conn.Close()
	logInfo("New client connection", s.conn.RemoteAddr())

	// configure client in queue
	clientQueue.setClient(s.conn)
	defer clientQueue.setClient(nil)

	// enable client
	s.clientActive = true

	// if push accounts is enabled, send list of accounts to client
	if conf.PushAccounts {
		s.handleAccountList()
	}

	// start client command handling loop
	r := bufio.NewReader(s.conn)
	c := ""
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
			s.handleCommand(c[:len(c)-2])
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
	defer l.Close()
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

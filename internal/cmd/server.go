package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"
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
	w := bufio.NewWriter(s.conn)
	n, err := w.WriteString(msg)
	if n < len(msg) || err != nil {
		s.clientActive = false
		log.Println("send:", len(msg), n, err)
	}
	if err := w.Flush(); err != nil {
		s.clientActive = false
		log.Println("send flush:", err)
	}
}

// handleAccountList handles an account list command
func (s *server) handleAccountList() {
	for _, a := range getAccounts() {
		// send replies with the following format:
		// account: <id> <name> <protocol> <user> <status>
		r := fmt.Sprintf("account: %d %s %s %s %s\r\n", a.ID, "()",
			a.Protocol, a.User, "online")
		log.Println(r)
		s.sendClient(r)
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
		log.Println("add NYI")
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

	// handle other commands
	switch parts[2] {
	case "delete":
		log.Println("delete", id, "NYI")
	case "buddies":
		log.Println("buddies", id, "NYI")
	case "collect":
		log.Println("collect", id, "NYI")
	case "send":
		log.Println("send", id, "NYI")
	case "status":
		log.Println("status", id, "NYI")
	case "chat":
		log.Println("chat", id, "NYI")
	}
}

// handleCommand handles a command received from the client
func (s *server) handleCommand(cmd string) {
	log.Println("client:", cmd)

	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "account":
		s.handleAccountCommand(parts)
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
	log.Println("New client connection", s.conn.RemoteAddr())

	s.clientActive = true
	r := bufio.NewReader(s.conn)
	c := ""
	for s.clientActive {
		// read a cmd line from the client
		cmd, err := r.ReadString('\n')
		if err != nil {
			log.Println("client:", err)
			return
		}

		// read and concatenate cmd lines until "\r\n"
		c += cmd
		if len(c) > 2 && c[len(c)-2] == '\r' {
			s.handleCommand(c[:len(c)-2])
			c = ""
		}
	}
}

// run runs the server
func (s *server) run() {
	// start listener
	l, err := net.Listen(s.network, s.address)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	s.listener = l
	s.serverActive = true

	// handle client connections
	log.Println("Server waiting for client connection")
	for s.serverActive {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println(err)
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
		network: "tcp",
		address: "localhost:32000",
	}
	server.run()
}

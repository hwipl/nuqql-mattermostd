package cmd

import (
	"bufio"
	"log"
	"net"
)

// server stores server information
type server struct {
	network  string
	address  string
	listener net.Listener
	conn     net.Conn
}

// handleCommand handles a command receviced from the client
func (s *server) handleCommand(cmd string) {
	log.Println("client:", cmd)
}

// handleClient handles a single client connection
func (s *server) handleClient() {
	defer s.conn.Close()
	log.Println("New client connection", s.conn.RemoteAddr())

	r := bufio.NewReader(s.conn)
	c := ""
	for {
		// read a cmd line from the client
		cmd, err := r.ReadString('\n')
		if err != nil {
			log.Println("client:", err)
			return
		}

		// read and concatenate cmd lines until "\r\n"
		c += cmd
		if len(c) > 2 && c[len(c)-2] == '\r' {
			s.handleCommand(c)
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

	// handle client connections
	log.Println("Server waiting for client connection")
	for {
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

package cmd

import (
	"bufio"
	"log"
	"net"
)

var (
	clientQueue = newQueue()
)

// queue stores messages for the client
type queue struct {
	queue    []string
	client   net.Conn
	clients  chan net.Conn
	messages chan string

	noHistory bool
	history   []string
	histReqs  chan struct{}
}

// sendToClient sends the contents of the message queue to the client
func (q *queue) sendToClient() {
	w := bufio.NewWriter(q.client)
	for len(q.queue) > 0 {
		msg := q.queue[0]
		n, err := w.WriteString(msg)
		if n < len(msg) || err != nil {
			q.client.Close()
			q.client = nil
			log.Println(err)
			break
		}
		if err := w.Flush(); err != nil {
			q.client.Close()
			q.client = nil
			log.Println(err)
			break
		}
		q.queue = q.queue[1:]
	}
}

// sendHistoryToClient sends the contents of the message history to the client
func (q *queue) sendHistoryToClient() {
	w := bufio.NewWriter(q.client)
	for _, msg := range q.history {
		n, err := w.WriteString(msg)
		if n < len(msg) || err != nil {
			q.client.Close()
			q.client = nil
			log.Println(err)
			break
		}
		if err := w.Flush(); err != nil {
			q.client.Close()
			q.client = nil
			log.Println(err)
			break
		}
	}
}

// run starts the main loop of the queue
func (q *queue) run() {
	for {
		select {
		case c, more := <-q.clients:
			// handle client (de)registration
			if !more {
				q.clients = nil
			}
			q.client = c

			if q.client != nil {
				// new client, send all queued messages to
				// the client
				q.sendToClient()
			}

		case m, more := <-q.messages:
			// handle message for client
			if !more {
				q.messages = nil
			}

			// append message to the message queue
			q.queue = append(q.queue, m)

			// add message to the history
			if !q.noHistory {
				q.history = append(q.history, m)
			}

			// send all queued messages to client if it's active
			if q.client == nil {
				continue
			}
			q.sendToClient()

		case _, more := <-q.histReqs:
			// handle get history request from client
			if !more {
				q.histReqs = nil
			}

			// send all messages in history to client
			if q.client == nil {
				continue
			}
			q.sendHistoryToClient()
		}

		// all channels closed, stop here
		if q.clients == nil && q.messages == nil && q.histReqs == nil {
			return
		}
	}
}

// getHistory requests the complete message history from the queue
func (q *queue) getHistory() {
	q.histReqs <- struct{}{}
}

// send sends msg to the (future) client via the queue
func (q *queue) send(msg string) {
	q.messages <- msg
}

// setClient sets conn as the client
func (q *queue) setClient(conn net.Conn) {
	q.clients <- conn
}

// newQueue creates a new queue
func newQueue() *queue {
	q := queue{
		clients:  make(chan net.Conn),
		messages: make(chan string),
		histReqs: make(chan struct{}),
	}
	go q.run()
	return &q
}

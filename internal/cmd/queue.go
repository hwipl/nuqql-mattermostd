package cmd

import (
	"bufio"
	"net"
)

var (
	clientQueue *queue
)

// queue stores messages for the client
type queue struct {
	queue    []string
	client   net.Conn
	clients  chan net.Conn
	messages chan string
}

// sendToClient sends the contents of the message queue to the client
func (q *queue) sendToClient() {
	w := bufio.NewWriter(q.client)
	for len(q.queue) > 0 {
		msg := q.queue[0]
		n, err := w.WriteString(msg)
		if n < len(msg) || err != nil {
			if err := q.client.Close(); err != nil {
				logError(err)
			}
			q.client = nil
			logError(err)
			break
		}
		if err := w.Flush(); err != nil {
			if err := q.client.Close(); err != nil {
				logError(err)
			}
			q.client = nil
			logError(err)
			break
		}
		q.queue = q.queue[1:]
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

			// send all queued messages to client if it's active
			if q.client == nil {
				continue
			}
			q.sendToClient()
		}

		// all channels closed, stop here
		if q.clients == nil && q.messages == nil {
			return
		}
	}
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
	}
	go q.run()
	return &q
}

// initClientQueue initializes the client queue
func initClientQueue() {
	clientQueue = newQueue()
}

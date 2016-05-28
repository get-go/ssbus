package main

import (
	"fmt"
	"io"
)

// A Bus holds open client connections, listens for incoming events on its
// Notifier channel, and broadcast event data to all registered connections.
type Bus struct {

	// Prefix for the server status messages
	Prefix string

	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

// Listen on different channels and act accordingly
func (bus *Bus) Listen(w io.Writer) {
	go func() {
		for {
			select {
			case s := <-bus.newClients:
				// A new client has connected, register their message channel
				bus.clients[s] = true

				if w != nil {
					fmt.Fprintf(w, "%s Client Added\n", bus.Prefix)
				}

			case s := <-bus.closingClients:
				// A client has dettached and we want to stop sending them messages.
				delete(bus.clients, s)

				if w != nil {
					fmt.Fprintf(w, "%s Client Removed\n", bus.Prefix)
				}

			case msg := <-bus.Notifier:
				// We got a new message, send message to all connected clients
				for clientMessageChan := range bus.clients {
					clientMessageChan <- msg
				}

				if w != nil {
					fmt.Fprintln(w, string(msg))
				}
			}
		}
	}()
}

func (bus *Bus) Write(b []byte) (int, error) {
	bus.Notifier <- b
	return len(b), nil
}

// New Bus object, already listening
func New() *Bus {
	return &Bus{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
		Prefix:         "[ssbus]",
	}
}

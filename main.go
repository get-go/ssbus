package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
)

// A Bus holds open client connections,
// listens for incoming events on its Notifier channel,
// and broadcast event data to all registered connections
type Bus struct {

	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

func (bus *Bus) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if req.Method == "POST" {
		scanner := bufio.NewScanner(req.Body)
		defer req.Body.Close()

		//read through all lines in the reqest, and post EACH line as a message
		for scanner.Scan() {
			msg := scanner.Text()
			if len(msg) > 0 {
				fmt.Fprintln(os.Stdout, msg)
				err := os.Stdout.Sync()
				if err != nil {
					log.Fatal(err)
					return
				}

				bus.Notifier <- []byte(msg)
				fmt.Fprintln(rw, msg)
			}
		}
		return
	}

	// Make sure that the writer supports flushing.
	flusher, ok := rw.(http.Flusher)

	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set the headers related to event streaming.
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Bus's connections registry
	messageChan := make(chan []byte)

	// Signal the bus that we have a new connection
	bus.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		bus.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := rw.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		bus.closingClients <- messageChan
	}()

	// block waiting for messages broadcast on this connection's messageChan
	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		fmt.Fprintln(rw, string(<-messageChan))

		// Flush the data immediatly instead of buffering it for later.
		flusher.Flush()
	}
}

// NewServer factory
func NewServer() (bus *Bus) {
	// Instantiate a bus
	bus = &Bus{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Set it running - listening and broadcasting events
	go bus.listen()

	return
}

// Listen on different channels and act accordingly
func (bus *Bus) listen() {
	for {
		select {
		case s := <-bus.newClients:
			// A new client has connected.
			// Register their message channel
			bus.clients[s] = true

		case s := <-bus.closingClients:
			// A client has dettached and we want to
			// stop sending them messages.
			delete(bus.clients, s)

		case event := <-bus.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan := range bus.clients {
				clientMessageChan <- event
			}
		}
	}
}

func catchClose() {
	// Watch for signal to close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				fmt.Fprintln(os.Stdout, "\nInterrupt Signal caught, exiting.")
				os.Exit(0)
			}
		}
	}()
}

func watchStdin(bus *Bus) {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := scanner.Text()
			if len(msg) > 0 {
				for clientMessageChan := range bus.clients {
					clientMessageChan <- []byte(msg)
				}
			}
		}
	}()
}

func main() {
	bus := NewServer()

	catchClose()

	watchStdin(bus)

	log.Fatal("HTTP server error: ", http.ListenAndServe("localhost:8080", bus))

}

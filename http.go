package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

// ListenAndServe implements 'http.ListenAndServe'
func (bus *Bus) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, bus)
}

func (bus *Bus) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	if !strings.HasPrefix(req.URL.Path, "/_") {
		fmt.Fprintf(bus, "Bad request. Path: %s", req.URL.Path)
		http.NotFound(rw, req)
		return
	}

	if req.Method == "POST" {
		// Post a message to the channels
		scanner := bufio.NewScanner(req.Body)
		defer req.Body.Close()

		// Read through all lines in the reqest, and post EACH line as a message
		for scanner.Scan() {
			msg := scanner.Text()
			if len(msg) > 0 {
				bus.Notifier <- []byte(msg)
				fmt.Fprintf(rw, "Success! Message sent: \"%s\"", msg)
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

	// Remove this client from the map of connected clients when this handler exits.
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
		// Write to the ResponseWriter 'Server Sent Events' compatible
		fmt.Fprintln(rw, string(<-messageChan))

		// Flush the data immediatly instead of buffering it for later
		flusher.Flush()
	}
}

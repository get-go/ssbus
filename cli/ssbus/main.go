package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/get-go/ssbus"
)

var stdin = flag.Bool("stdin", false, "Accept input on Standard In")
var quiet = flag.Bool("quiet", false, "Quiet down the Standard Out messages")
var addr = flag.String("address", ":8675", "Address to listen on")
var logFile = flag.String("logfile", "", "File to save logs to")
var prefix = flag.String("prefix", "[ssbus]", "Prefix for server messages")

func main() {
	flag.Parse()

	//Make us a new bus
	bus := ssbus.New()

	//Watch for a close 'os.Signal'
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		//Throw in go func to watch continuously
		for sig := range c {
			if sig == os.Interrupt {
				fmt.Fprintf(bus, "%s Interrupt Signal caught, exiting.\n", bus.Prefix)
				os.Exit(0)
			}
		}
	}()

	// log will be used to write log output to; file, system io, stream
	var log io.Writer

	if *logFile != "" {
		// logFile get's priority
		log, _ = os.Open(*logFile)
	} else if !*quiet {
		// standard out by default
		log = os.Stdout
	} else {
		// keep it secret, keep it safe
		log = nil
	}

	// start listening on the bus, log can be nil
	bus.Listen(log)

	if *stdin {
		// watch for lines incoming on 'os.Stdin'
		go func() {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				msg := scanner.Text()
				if len(msg) > 0 {
					bus.Notifier <- []byte(msg)
				}
			}
		}()
	}

	//Send initial status messages
	fmt.Fprintf(bus, "%s Starting Stupid Simple Bus service\n", bus.Prefix)
	defer fmt.Fprintf(bus, "%s Stopping Stupid Simple Bus service\n", bus.Prefix)

	//Start an HTTP server, on the specified address
	//This is blocking, and will return an error when done
	err := bus.ListenAndServe(":8675")
	if err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %+v\n", err)
		os.Exit(1)
	}
}

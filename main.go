// This program implements a simple example of a Slack bot.  It has the capabiltiy
// of taking the user to the Github site for this code plus it can show you a
// picture of a certain dog.  Hey, it's my first bot ever!
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	portNum int // listen port
	timeout int // server timeout in seconds
)

func init() {
	flag.IntVar(&portNum, "port", 80, "HTTP port number")
	flag.IntVar(&timeout, "timeout", 30, "server timeout (seconds)")
}

func main() {
	flag.Parse()

	lg := log.New(os.Stdout, "ggbot: ", log.Lshortfile|log.LstdFlags)
	ctx := context.Background()

	// Create the Slack event listener
	s, err := NewSlack(lg)
	if err != nil {
		lg.Fatal(err)
	}

	// And run the event loop.
	if err := s.Run(ctx); err != nil {
		lg.Fatal(err)
	}

	// Since we are usering Interactive Components for callbacks,
	// we need an HTTP handler.
	handler, err := NewHandler()
	if err != nil {
		lg.Fatal(err)
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         fmt.Sprintf(":%d", portNum),
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}

	// Start the server.
	go func() {
		log.Println("[INFO] Listening for connections", "port", portNum)
		if err := srv.ListenAndServe(); err != nil {
			log.Println("[INFO] Server completed", "err", err)
		}
	}()

	// Block until we shutdown.
	waitForShutdown(ctx, srv, lg)
}

// Setup for clean shutdown with signal handlers/cancel.
func waitForShutdown(ctx context.Context, srv *http.Server,
	log *log.Logger) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	sig := <-interruptChan
	log.Println("[INFO] Termination signal received", "signal", sig)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)
	log.Printf("[INFO] Shutting down")
}

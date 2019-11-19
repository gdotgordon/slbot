// This program implements a simple example of a Slack bot.
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

	s, err := NewPost(lg)
	if err != nil {
		lg.Fatal(err)
	}

	if err := s.Run(ctx); err != nil {
		lg.Fatal(err)
	}

	handler, err := s.NewHandler()
	if err != nil {
		lg.Fatal(err)
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         fmt.Sprintf(":%d", portNum),
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}

	// Start server
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

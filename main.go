package main

import (
	"context"
	"log"
	"net/http"
	"os"
)

func main() {
	lg := log.New(os.Stdout, "ggbot: ", log.Lshortfile|log.LstdFlags)
	ctx := context.Background()

	s, err := NewPost()
	if err != nil {
		lg.Fatal(err)
	}
	s.Logger = lg

	if err := s.Run(ctx); err != nil {
		s.Logger.Fatal(err)
	}

	handler, err := s.NewHandler()
	if err != nil {
		s.Logger.Fatal(err)
	}
	lg.Fatal(http.ListenAndServe(":9191", handler))
}

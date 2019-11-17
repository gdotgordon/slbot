package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/nlopes/slack"
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("incoming url: ", r.URL)
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("incorrect path: %s", r.URL.Path)))
		return
	}

	if r.Body == nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("empty body"))
		return
	}
	defer r.Body.Close()

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not parse body"))
		return
	}

	// slack API calls the data POST a 'payload'
	reply := r.PostFormValue("payload")
	if len(reply) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("could not find payload"))
		return
	}

	var payload slack.InteractionCallback
	err = json.NewDecoder(strings.NewReader(reply)).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte("could not process payload"))
		return
	}

	//action := payload.Actions[0].Value
	action := payload.ActionCallback.AttachmentActions[0]
	fmt.Printf("Action: %+v\n", action)
	switch action.Value {
	case "yes":
		w.Write([]byte("The source code is at https://github.com/gdotgordon/slbot!"))
	case "no":
		w.Write([]byte("No worries, let me know later on if you do!"))
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("could not process callback: %s", action.Value)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// askIntent is the initial request back to user if they'd like to see
// the scores from the most recent slate of games
//
// NOTE: This is a contrived example of the functionality, but ideally here
// we would ask users to specify a date, or maybe a team, or even
// a specific game which we could present back
func (s *Slack) askIntent(ev *slack.MessageEvent) error {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(middleware.NoCache)
	r.Use(middleware.Heartbeat("/ping"))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Would you like to see the source code for this bot?",
		CallbackID: fmt.Sprintf("ask_%s", ev.User),
		Color:      "#666666",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "action",
				Text:  "No thanks!",
				Type:  "button",
				Value: "no",
			},
			slack.AttachmentAction{
				Name:  "action",
				Text:  "Yes, please!",
				Type:  "button",
				Value: "yes",
			},
		},
	}

	params.User = ev.User
	params.AsUser = true

	_, err := s.Client.PostEphemeral(
		ev.Channel,
		ev.User,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionPostMessageParameters(params),
	)
	if err != nil {
		return err
	}

	return nil
}

// NewHandler instantiaties the web handler for listening on the API
func (s *Slack) NewHandler() (http.Handler, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	r.Use(middleware.NoCache)
	r.Use(middleware.Heartbeat("/ping"))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)
	r.Post("/", postHandler)

	r.Get("/debug/pprof/*", pprof.Index)
	r.Get("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		first := true
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		expvar.Do(func(kv expvar.KeyValue) {
			if !first {
				fmt.Fprintf(w, ",\n")
			}
			first = false
			fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
		})
		fmt.Fprintf(w, "\n}\n")
	})

	return r, nil
}

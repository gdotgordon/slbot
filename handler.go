// The handler is the HTTP callback from Slack at the registered URL after
// the uer has chosen some action.
//
// This is my modifcation of the sample in Gopher Academy at
// https://github.com/sebito91/nhlslackbot, so here's the
// required copyright:
//
//Copyright (c) 2017 Sebastian Borza
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	/*
		"github.com/go-chi/chi"
		"github.com/go-chi/chi/middleware"
		"github.com/go-chi/cors"
	*/
	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

func postHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("********incoming url: ", r.URL)
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
		attachment := slack.Attachment{
			Text:     "",
			Color:    "#334fff",
			ImageURL: "https://i.imgur.com/uVANlUI.jpg",
			Actions:  []slack.AttachmentAction{},
			Fields: []slack.AttachmentField{
				{
					Title: "A dog",
				},
			},
		}
		msg := payload.OriginalMessage
		msg.Attachments = []slack.Attachment{attachment}

		//w.Write([]byte("OK, but your missing something great!"))
		origMsg := payload.OriginalMessage
		fmt.Printf("original: %+v\n", origMsg)
		origMsg.Attachments = []slack.Attachment{
			{
				Text:    "",
				Actions: []slack.AttachmentAction{},
				Fields: []slack.AttachmentField{
					{
						Title: "A dog",
					},
				},
				ImageURL: "https://i.imgur.com/uVANlUI.jpg",
			},
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&msg)
		fmt.Println(b.String())
		w.Write(b.Bytes())
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("could not process callback: %s", action.Value)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// askIntent is the initial request back to user if they'd like to see
// the location of the repo
func (s *Slack) askIntent(ev *slack.MessageEvent, user *slack.User) error {
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Would you like to see the source code for this bot?",
		CallbackID: fmt.Sprintf("ask_%s", ev.User),
		Color:      "#334fff",
		Actions: []slack.AttachmentAction{
			slack.AttachmentAction{
				Name:  "action",
				Text:  "Yes, I love Go code!",
				Type:  "button",
				Value: "yes",
			},
			slack.AttachmentAction{
				Name:  "action",
				Text:  "No thanks, boring.",
				Type:  "button",
				Value: "no",
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
	r := mux.NewRouter()
	r.HandleFunc("/", postHandler).Methods(http.MethodPost)
	return r, nil
}

// The handler is the HTTP callback from Slack at the registered URL after
// the uer has chosen some action.
//
// This is my significantly modified version of the sample at Gopher Academy:
// https://github.com/sebito91/nhlslackbot.  As it is an MIT licesne, here is
// the required attribution:
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

	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

// NewHandler instantiaties the web handler for listening on the API
func NewHandler() (http.Handler, error) {
	r := mux.NewRouter()
	r.HandleFunc("/", postHandler).Methods(http.MethodPost)
	return r, nil
}

func postHandler(w http.ResponseWriter, r *http.Request) {
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

	action := payload.ActionCallback.AttachmentActions[0]
	msg := payload.OriginalMessage

	//fmt.Printf("********got action value: %+v\n", action)
	value := action.Value
	if value == "" {
		value = action.Name
	}
	switch value {
	case "yesMe":
		showMe(w, msg)
		return
	case "noMe":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK, but I assure you I'm a nice guy!"))
		return
	case "noCode":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK, I guess you prefer Java!"))
		return
	case "actionSelect":
		switch action.SelectedOptions[0].Value {
		case "Me":
			showMe(w, msg)
			return
		case "Code":
			// Overwrite original drop down message.
			msg.Attachments = []slack.Attachment{
				{
					Text:  "Would you like to see the source code for this bot, including instructions?",
					Color: "#334fff",
					Actions: []slack.AttachmentAction{
						slack.AttachmentAction{
							Name:  "action",
							Text:  "Yes, I love Go code!",
							Type:  "button",
							Value: "yes_code",
							URL:   "https://github.com/gdotgordon/slbot",
						},
						slack.AttachmentAction{
							Name:  "action",
							Text:  "No thanks, boring.",
							Type:  "button",
							Value: "noCode",
						},
					},
				},
			}

			w.Header().Add("Content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&msg)
			return
		}
	case "actionCancel":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hope to hear from you again soon!"))
		return
	default:
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("could not process callback: %s", action.Value)))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func showMe(w http.ResponseWriter, msg slack.Message) {
	attachment := slack.Attachment{
		Text:     "",
		Color:    "#334fff",
		ImageURL: "https://i.imgur.com/VlU0uLt.jpg",
		Actions:  []slack.AttachmentAction{},
		Fields: []slack.AttachmentField{
			{
				Title: "Me",
			},
		},
	}
	msg.Attachments = []slack.Attachment{attachment}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(&msg)
	//fmt.Println(b.String())
	w.Write(b.Bytes())
}

// This is the slack event listener code for when the user triggers
// an action.
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
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"
)

// Slack is the primary struct for our slackbot
type Slack struct {
	Name  string
	Token string

	User   string
	UserID string

	Logger *log.Logger

	Client *slack.Client
}

// NewPost returns a new instance of the Slack struct, primary for our slackbot
func NewPost(log *log.Logger) (*Slack, error) {
	token := os.Getenv("SLACK_TOKEN")
	if len(token) == 0 {
		return nil, fmt.Errorf("could not discover API token")
	}

	return &Slack{
		Client: slack.New(token, slack.OptionLog(log)),
		Token:  token, Name: "ggbot", Logger: log}, nil
}

// Run is the primary service to generate and kick off the slackbot listener
// This portion receives all incoming Real Time Messages notices from the workspace
// as registered by the API token
func (s *Slack) Run(ctx context.Context) error {
	authTest, err := s.Client.AuthTest()
	if err != nil {
		return fmt.Errorf("did not authenticate: %+v", err)
	}

	s.User = authTest.User
	s.UserID = authTest.UserID

	s.Logger.Printf("[INFO]  bot is now registered as %s (%s)\n", s.User, s.UserID)

	go s.run(ctx)
	return nil
}

func (s *Slack) run(ctx context.Context) {
	rtm := s.Client.NewRTM()
	go rtm.ManageConnection()

	s.Logger.Printf("[INFO]  now listening for incoming messages...")
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.User) == 0 {
				continue
			}

			fmt.Println("message text", ev.Msg.Text)

			// check if we have a DM, or standard channel post
			direct := strings.HasPrefix(ev.Msg.Channel, "D")

			if !direct && !strings.Contains(ev.Msg.Text, "@"+s.UserID) {
				// msg not for us!
				continue
			}

			//fmt.Printf("event: %+v\n", ev)
			user, err := s.Client.GetUserInfo(ev.User)
			if err != nil {
				s.Logger.Printf("[WARN]  could not grab user information: %s", ev.User)
				continue
			}
			//fmt.Printf("user: %+v\n", user)

			s.Logger.Printf("[DEBUG] received message from %s (%s)\n", user.Profile.RealName, ev.User)

			err = s.askIntent(ev, user)
			if err != nil {
				s.Logger.Printf("[ERROR] posting ephemeral reply to user (%s): %+v\n", ev.User, err)
			}
		case *slack.RTMError:
			s.Logger.Printf("[ERROR] %s\n", ev.Error())
		}
	}
}

// This is the slack event listener code for when the user triggers
// an action.
//
// This is my significantly modified version of the sample at Gopher Academy:
// https://github.com/sebito91/nhlslackbot.  As it is an MIT licesne, here is
// the required attribution:
//
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

// NewSlack returns a new instance of the Slack struct, primarily for our slackbot
func NewSlack(log *log.Logger) (*Slack, error) {
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

// run is the main run loop for grabbing slack events
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

			//fmt.Println("message text", ev.Msg.Text)

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
			s.Logger.Printf("[DEBUG] received message from %s (%s)\n", user.Profile.RealName, ev.User)

			if strings.Contains(ev.Msg.Text, "code") {
				err = s.askCodeIntent(ev, user)
				if err != nil {
					s.Logger.Printf("[ERROR] posting ephemeral reply to user (%s): %+v\n", ev.User, err)
				}
			} else if strings.Contains(ev.Msg.Text, "dog") {
				err = s.askDogIntent(ev, user)
				if err != nil {
					s.Logger.Printf("[ERROR] posting ephemeral reply to user (%s): %+v\n", ev.User, err)
				}
			} else {
				err = s.askGeneralIntent(ev, user)
				if err != nil {
					s.Logger.Printf("[ERROR] posting ephemeral reply to user (%s): %+v\n", ev.User, err)
				}
			}
		case *slack.RTMError:
			s.Logger.Printf("[ERROR] %s\n", ev.Error())
		}
	}
}

// askCodeIntent is the initial request back to user if they'd like to be taken to
// the git repo for this code.
func (s *Slack) askCodeIntent(ev *slack.MessageEvent, user *slack.User) error {
	params := slack.PostMessageParameters{}
	attachment := codeAttachment(ev.User)
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

// askDogIntent is the initial request back to user if they'd like to see
// the location of the repo
func (s *Slack) askDogIntent(ev *slack.MessageEvent, user *slack.User) error {
	params := slack.PostMessageParameters{}
	attachment := dogAttachment(ev.User)

	params.User = ev.User
	params.AsUser = true
	_, _, err := s.Client.PostMessage(ev.Channel, slack.MsgOptionAttachments(attachment), slack.MsgOptionPostMessageParameters(params))
	return err
}

// askGeneralIntent is the initial request back to user if they haven't specifed
// a particular supported keyword (dog or code).  It puts up a menu of choices.
func (s *Slack) askGeneralIntent(ev *slack.MessageEvent, user *slack.User) error {
	params := slack.PostMessageParameters{}
	attachment := slack.Attachment{
		Text:       "Please select one of the supported operations:",
		CallbackID: fmt.Sprintf("ask_%s", ev.User),
		Color:      "#334fff",
		Actions: []slack.AttachmentAction{
			{
				Value: "actionSelect",
				Name:  "actionSelect",
				Type:  "select",
				Options: []slack.AttachmentActionOption{
					{
						Text:  "Go to Github to See This Code",
						Value: "Code",
					},
					{
						Text:  "See a Cute Dog",
						Value: "Dog",
					},
				},
			},

			{
				Name:  "actionCancel",
				Text:  "Cancel",
				Type:  "button",
				Style: "danger",
				Value: "actionCancel",
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

func codeAttachment(user string) slack.Attachment {
	return slack.Attachment{
		Text:       "Would you like to see the source code for this bot, including instructions?",
		CallbackID: fmt.Sprintf("ask_%s", user),
		Color:      "#334fff",
		Actions: []slack.AttachmentAction{
			{
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
	}
}

func dogAttachment(user string) slack.Attachment {
	return slack.Attachment{
		Text:       "Would you like to see a picture of a dog?",
		CallbackID: fmt.Sprintf("ask_%s", user),
		Color:      "#334fff",
		Actions: []slack.AttachmentAction{
			{
				Name:  "yesDog",
				Text:  "Yes, I love dogs!",
				Type:  "button",
				Value: "yesDog",
			},
			{
				Name:  "noDog",
				Text:  "No thanks, I prefer cats.",
				Type:  "button",
				Value: "noDog",
			},
		},
	}
}

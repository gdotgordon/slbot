# slbot
My first Slack bot

## What Does This Do?
In an effort the learn the Slack API, in particular in the context of Slack bots, I set out to write a Slack app and bind it to a bot.  When I say "bind it to a bot", I mean the bot will be triggered either when you DM (direct message) it, or when you "cite" it via the "@", construct.  In this case the bot is called "ggbot", so you can simply start a conversation with the user "ggbot" and type things containing the word "code" (to get directed to this Github site to see the code) or "picture" (too see my picture), or just type anything to get a list of all the options.  Alternatively go into any of the regular channels where ggbot has joined (such as "general") and type a line like ```@ggbot``` or ```@ggbot code``` ```@ggbot show me the picture.```

Note, this is very much a work in progress.  The basic functionality is there, but more cleanup is needed with regard to so-called "ephemeral" messages vs. regular messages.  But it works well enough to demonstrate the concepts.
The architecture is basically a Slack event listener coupled with an HTTP handler, which is required when the app is configured as an "interactive component".

## How Do I Use It
I've pretty much already described how this works.  The initial functionality provides two "services" - taking the user to this Github site to view the source code, and also to show a picture of me.  There are two ways you can activate the bot:

- Direct Message ggbot: start or resume a DM to "ggbot".  If you ask something with one of the two keywords "code" or "picture", it will ask you if you want to proceed with those, and will take you to the code or show the picture if you do.  If you type anything else, it will give you a pop-up with the two choices.

- Go to a channel "ggbot has joined (such as "general") and type`@ggbot` plus a space (or anything to make the reference active), then send the message.  You'll get a dropdown menu with two options: see a picture of me, or open a browser to this github repo.  Regardless of which one you choose, you'll then get a Yes/No confirmation to proceed.  Click the button of your choice.  Or you can address it with words containing "picture" or "code", as in `@ggbot picture` or `@ggbot Go code is cool.` to have it ask if you want that specific action done.

Enjoy!

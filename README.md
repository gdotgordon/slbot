# slbot
My first Slack bot

## What Does This Do?
In an effort the learn the Slack API, in particlar in the context of Slack bots, I set up to write a Slack app and bind it to a bot.  When I say "bind it to a bot", I mean the bot will be triggered when you "cite" it via the "@", constuct.  In this case the bot is called "ggbot", so you can go into any of the regular channels where it has joined and type a line like ```@ggbot``` or ```gbot code``` (to get directed to this Github site to see the code).

Note, this is very much a work in progress.  The basic functionality is there, but more cleanup is needed with regard to so-called "ephemeral" messages vs. regular messages.  But it works well enough to demonstrate the concepts.
The arhcitecture is basically a Slack event listener coupled with an HTTP handler, which is required when the app is configured as an "interactive component".

## How Do I Use It
The initial functionality provides two "services": taking the user to this Github site to view the source code, and also to show a picture of me.  You can either simply invoke the bot by typing `@ggbot` plus a space to make the link active, then send ther message.  You'll get a dropdown menu with two options: see a picture of me, or open a browser to this github repo.  Regardless of which one you choose, you'll then get a Yes/No confirmaton to proceed.  Click the button of your choice.

The other way to use the bot is to address it with words containing "picture" or "code", as in `@ggbot picture` or `@ggbot code is cool`.  In this 

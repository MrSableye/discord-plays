# pokemon-bot

<p float="center">
  <img src="/screenshot.png" height="250" />
  <img src="/screenshot2.png" height="250" />
  <img src="/screenshot3.png" height="250" />
</p>

## Commands:
Movement: `/u /d /l /r`, `up down left right`    
Action buttons: `/a /b /start /select`    

Commands can be used with a count to execute them many times:
`/l 10`    

Maximum value: 10

Useful commands:    
`/party-count` prints party count and pokemon info    
`/ball-count` prints ball count    
`/trainer` prints general info    
`/save` saves the game through the in game dialog and dumps the battery    
`/map` opens the map, takes a screenshot and returns it to the bot, exits map    
`/summary` shows a gif of the last few frames    
`/leaderboard` shows who pressed the most keys     

## Instructions:
To add this bot to your server, you need a bot token. Create an application and a bot, get the token and put it in the same directory as main.go as `token.txt`.

Compile `GameboyWebserver` with cmake    
Run: `./GameboyWebserver ./LegallyObtainedGameWithMonsters.gbc` (backend & emulator)    
then run `go run main.go` (discord bot)    

In case `main.go` crashes, report an issue with the log. Your progress is not lost, run it again.    

Worst case scenario, visit `http://localhost:1234/req?action=save` which saves a state and can be loaded with `http://localhost:1234/req?action=load`.

But usually you want to save with the `/save` command, and the battery save gets auto loaded on startup    

TODO: Make script that asks for token, rom etc. and makes run script

## Consists of 3 parts:

Go front-end, hosts the discord bot and communicates with the middle-end    
C++ middle-end, hosts an http server and communicates with the back-end    
C++ back-end, runs the gameboy emulator code    

`main.go` is the discord bot frontend that communicates with the C++ backend `main.cxx`    
`main.cxx` hosts a simple http server allowing admin control and intra process communication with `main.go`    
`gb_headlesswrapper.cxx` is the emulator wrapper that runs the emulator code    
`GameboyTKP/` is the actual emulator code    

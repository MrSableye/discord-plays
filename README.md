# pokemon-bot

## Commands:
Movement: `/u /d /l /r`, `up down left right`    
Action buttons: `/a /b /start /select`    

Useful commands:    
`/party-count` prints party count and pokemon info    
`/ball-count` prints ball count    
`/trainer` prints general info    

## Install:
TODO: Make script that asks for token, rom etc. and makes run script

## Consists of 3 parts:

Go front-end, hosts the discord bot and communicates with the middle-end    
C++ middle-end, hosts an http server and communicates with the back-end    
C++ back-end, runs the gameboy emulator code    

`main.go` is the discord bot frontend that communicates with the C++ backend `main.cxx`    
`main.cxx` hosts a simple http server allowing admin control and intra process communication with `main.go`    
`gb_headlesswrapper.cxx` is the emulator wrapper that runs the emulator code    
`GameboyTKP/` is the actual emulator code    

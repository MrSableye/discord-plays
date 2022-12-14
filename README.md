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
To add this bot to your server, follow these instructions:

- Compile `GameboyWebserver` with CMake. Take note of where the executable is stored
    - `cmake --build build`
    - `cmake -B build`
- While in directory `DiscordBot`, run the command `go run .`     
- Using the console menu, configure the bot. Make sure the generated config.json file is correct    
- Run the bot!
- You can run the bot in the background with `nohup go run . &` and exit the terminal, the bot won't exit.

## Consists of 3 parts:

Go front-end, hosts the discord bot and communicates with the middle-end through GET requests    
C++ middle-end, hosts an http server and communicates with the back-end    
C++ back-end, runs the gameboy emulator code    

`DiscordBot/` contains the discord bot code    
`main.cxx` hosts a simple http server allowing admin control and intra process communication with `main.go`    
`gb_headlesswrapper.cxx` is the emulator wrapper that runs the emulator code    
`GameboyTKP/` is the actual emulator code    

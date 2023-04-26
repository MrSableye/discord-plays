# pokemon-bot 

A discord bot hosting a GB/GBC/GBA emulator!

<p float="center">
  <img src="/screenshot.png" height="250" />
  <img src="/screenshot2.png" height="250" />
</p>

## Commands:

`/screen` shows the in-game screen with buttons    
`/summary` shows a gif of the last few frames (TODO)    
`/leaderboard` shows who pressed the most keys     
`/poke-jail` shows who's currently banned    

## Instructions:
To add this bot to your server, follow these instructions:

- Download (or compile) the latest version of SkyEmu
  - Windows: https://nightly.link/skylersaleh/SkyEmu/workflows/deploy_win/dev/WindowsRelease.zip
  - macOS: https://nightly.link/skylersaleh/SkyEmu/workflows/deploy_mac/dev/MacOSRelease.zip
  - Linux: https://nightly.link/skylersaleh/SkyEmu/workflows/deploy_linux/dev/LinuxRelease.zip
- While in directory `DiscordBot`, run the command `go build -o DiscordBot`     
- This should create an executable named `DiscordBot`, which you can run using the terminal
  - On Windows, open `cmd.exe`, navigate to the directory of the executable and run it by typing `DiscordBot.exe`
  - On Linux, open your terminal emulator, navigate to the directory of the executable and run it by typing `./DiscordBot`
- The first time you run the bot, it will help you create a configuration file    
  - Make sure you use the token, not the application id or the public key!
  <p float="center">
    <img src="/token.png"/>
  </p>
- Run the bot!

You should be able to use your own emulators if you follow the [SkyEmu API](https://github.com/skylersaleh/SkyEmu/blob/dev/docs/HTTP_CONTROL_SERVER.md)

## Administration
Create an `admins.json` file in `DiscordBot/` with a json string array of all user ids you want to be administrators of the bot.    
Example: `["21318712398012", "19238129031092"]` will have 2 admins with those ids.    

### Admin commands:
`/poke-ban <id> <opt:days> <opt:reason>` bans a user. If no days are specified the ban is for 9999 days.    
`/poke-unban <id>` unbans a user    

Banned users can't use bot commands.    

## Ack
Thanks to Sky for [SkyEmu](https://github.com/skylersaleh/SkyEmu)    
and my friends at the emudev discord.

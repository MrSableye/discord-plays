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

- Compile the latest version of [SkyEmu](https://github.com/skylersaleh/SkyEmu)    
- While in directory `DiscordBot`, run the command `go run .`     
- Using the console menu, configure the bot. Make sure the generated config.json file is correct.    
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

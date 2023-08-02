package main

import "github.com/bwmarrin/discordgo"

func commandLoad(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	checkOk(get("load?path=" + executablePath + "/save.png"))
	str := S["loadedSave"]
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &str,
	})
}

package main

import "github.com/bwmarrin/discordgo"

func commandPokeUnban(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i) {
		return
	}
	optionMap := getOptions(i)
	b := removeBanned(optionMap["user-id"])
	if b {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: SR("userUnbanned", i),
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: SR("userNotBanned", i),
			},
		})
	}
}

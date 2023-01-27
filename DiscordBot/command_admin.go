package main

import (
	"github.com/bwmarrin/discordgo"
)

func commandPokeAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !mustAdmin(s, i) {
		return
	}
	optionMap := getOptions(i)
	id := optionMap["user-id"]
	admins = append(admins, id)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         SR("userAdmin", i),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

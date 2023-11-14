package main

import "github.com/bwmarrin/discordgo"

func commandLeaderboard(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isBanned(s, i) {
		return
	}
	ldr := printLeaderboard()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Leaderboard",
					Description: ldr,
				},
			},
		},
	})
}

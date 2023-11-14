package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func commandJail(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i) {
		return
	}
	var sb strings.Builder
	for i := 0; i < len(bannedPlayers); i++ {
		sb.WriteString("<@!" + bannedPlayers[i].Id + ">, Reason: ")
		if bannedPlayers[i].Reason == "" {
			sb.WriteString("No reason given")
		} else {
			sb.WriteString(bannedPlayers[i].Reason)
		}
		sb.WriteString("\nwas timed out on " + bannedPlayers[i].BanDate.Format("2006-01-02"))
		sb.WriteString(" by " + bannedPlayers[i].BannedBy)
		sb.WriteString("\nand will be free to use the bot on " + bannedPlayers[i].UnbanDate.Format("2006-01-02") + "\n")
		sb.WriteString("\n")
		if i != len(bannedPlayers)-1 {
			sb.WriteString("\n")
		}
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Jail",
					Description: sb.String(),
				},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

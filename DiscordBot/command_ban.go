package main

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func commandPokeBan(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i) {
		return
	}
	optionMap := getOptions(i)
	var bannedPlayer BannedPlayer
	bannedPlayer.Id = optionMap["user-id"]
	bannedPlayer.BanDate = time.Now()
	bannedPlayer.BannedBy = "<@!" + i.Member.User.ID + ">"
	if optionMap["days"] != "" {
		days, err := strconv.Atoi(optionMap["days"])
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: SR("banInvalidDays", i),
				},
			})
			return
		}
		bannedPlayer.UnbanDate = time.Now().AddDate(0, 0, days)
	} else {
		bannedPlayer.UnbanDate = time.Now().AddDate(0, 0, 9999)
	}
	if optionMap["reason"] != "" {
		bannedPlayer.Reason = optionMap["reason"]
	}
	b := addBanned(bannedPlayer)
	if b {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:         SR("userBanned", i),
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:         SR("userAlreadyBanned", i),
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		})
	}
}

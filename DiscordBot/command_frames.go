package main

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

func commandFrames(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if checkBanned(s, i) {
		return
	}
	if !checkChannel(s, i) {
		return
	}
	if !checkRole(s, i) {
		return
	}
	settings.FramesSteppedPressed = int(i.ApplicationCommandData().Options[0].IntValue())
	lastPressTime = time.Now()
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Frames set to " + strconv.Itoa(settings.FramesSteppedPressed),
		},
	})
	check(err)
}

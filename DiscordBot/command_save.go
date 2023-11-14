package main

import (
	"bytes"
	"compress/gzip"
	"os"

	"github.com/bwmarrin/discordgo"
)

func commandSave(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isBanned(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

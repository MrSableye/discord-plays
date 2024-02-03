package main

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
)

func commandSave(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !mustAdmin(s, i) {
		return
	}
	if checkBanned(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	checkOk(get("save?path=" + executablePath + "/save.png"))
	b, _ := ioutil.ReadFile(executablePath + "/save.png")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(b)
	gz.Close()
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Files: []*discordgo.File{
			{Name: "save.png.gz", Reader: &buf},
		},
	})
}

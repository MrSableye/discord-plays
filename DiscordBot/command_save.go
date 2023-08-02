package main

import (
	"bytes"
	"compress/gzip"
	"os"

	"github.com/bwmarrin/discordgo"
)

func commandSave(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if checkBanned(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	checkOk(get("save?path=" + executablePath + "/save.png"))
	b, _ := os.ReadFile(executablePath + "/save.png")
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

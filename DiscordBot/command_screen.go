package main

import (
	"bytes"
	"image/png"

	"github.com/bwmarrin/discordgo"
	"github.com/nfnt/resize"
)

func commandScreen(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if checkBanned(s, i) {
		return
	}
	if !checkChannel(s, i) {
		return
	}
	if !checkRole(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	reader := getScreen("png")
	img, err := png.Decode(reader)
	check(err)
	img = resize.Resize(settings.WidthOfImage, 0, img, resize.Lanczos3)
	writer := new(bytes.Buffer)
	err = png.Encode(writer, img)
	check(err)
	embeds := []*discordgo.MessageEmbed{
		{
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://screen.png",
			},
		},
	}
	buttons := getButtons()
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &embeds,
		Files: []*discordgo.File{
			{Name: "screen.png", Reader: writer},
		},
		Components: &buttons,
	})
}

package main

import (
	"github.com/bwmarrin/discordgo"
)

func commandScreen(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !isAdmin(s, i) {
		return
	}
	if checkBanned(s, i) {
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsHasThread,
		},
	})

	// TODO: bring back?
	// reader := getScreen("png")
	// img, err := png.Decode(reader)
	// check(err)
	// img = resize.Resize(settings.WidthOfImage, 0, img, resize.Lanczos3)
	// writer := new(bytes.Buffer)
	// err = png.Encode(writer, img)
	// check(err)
	// embeds := []*discordgo.MessageEmbed{
	// 	{
	// 		Image: &discordgo.MessageEmbedImage{
	// 			URL: "attachment://screen.png",
	// 		},
	// 		Footer: &discordgo.MessageEmbedFooter{
	// 			Text: "https://github.com/OFFTKP/pokemon-bot",
	// 		},
	// 	},
	// }
	buttons := getButtons()
	m, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		// Embeds: &embeds,
		// Files: []*discordgo.File{
		// 	{Name: "screen.png", Reader: writer},
		// },
		Components: &buttons,
	})
	check(err)

	_, err = s.MessageThreadStartComplex(i.ChannelID, m.ID, &discordgo.ThreadStart{
		Name:                "Chat",
		AutoArchiveDuration: 60 * 24 * 7,
		Invitable:           false,
		RateLimitPerUser:    0,
	})
	check(err)
}

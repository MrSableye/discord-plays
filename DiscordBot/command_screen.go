package main

import "github.com/bwmarrin/discordgo"

func commandScreen(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if checkBanned(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	reader := getScreen()
	embeds := []*discordgo.MessageEmbed{
		{
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://screen.png",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "https://github.com/OFFTKP/pokemon-bot",
			},
		},
	}
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &embeds,
		Files: []*discordgo.File{
			{Name: "screen.png", Reader: reader},
		},
		Components: &buttonComponents,
	})
}

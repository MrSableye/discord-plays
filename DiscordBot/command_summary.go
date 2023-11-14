package main

import (
	"github.com/bwmarrin/discordgo"
)

func commandSummary(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// isBanned(s, i)
	// s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
	// 	Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	// })
	// summaryMutex.Lock()
	// defer summaryMutex.Unlock()
	// get("gif")
	// data, err := os.ReadFile("out.gif")
	// if err != nil {
	// 	fmt.Println("Error while loading gif")
	// 	return
	// }
	// reader := bytes.NewReader(data)
	// embeds := []*discordgo.MessageEmbed{
	// 	{
	// 		Title: "GIF summary",
	// 		Image: &discordgo.MessageEmbedImage{
	// 			URL: "attachment://my.gif",
	// 		},
	// 		Footer: &discordgo.MessageEmbedFooter{
	// 			Text: "https://github.com/OFFTKP/pokemon-bot",
	// 		},
	// 	},
	// }
	str := "This command is currently disabled."
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &str,
		// Embeds: &embeds,
		// Files: []*discordgo.File{
		// 	{Name: "my.gif", Reader: reader},
		// },
	})
}

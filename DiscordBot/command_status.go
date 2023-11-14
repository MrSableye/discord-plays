package main

import (
	"fmt"
	"io/ioutil"

	"github.com/bwmarrin/discordgo"
)

func commandStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if isBanned(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	resp := get("status")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	str := "```c\n" + string(body) + "```"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &str,
	})
}

package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func stringToButton(s string) ButtonType {
	for i := 0; i < int(ButtonsCount); i++ {
		var button ButtonType = ButtonType(i)
		if button.String() == s {
			return button
		}
	}
	return -1
}

func commandHold(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if checkBanned(s, i) {
		return
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	optionMap := getOptions(i)

	var requestedButton = strings.ToLower(optionMap["button"])
	if requestedButton == "" {
		heldButtons = make([]ButtonType, 0)
		for i := 0; i < int(ButtonsCount); i++ {
			disabledButtons[i] = false
		}
		response := "All buttons released."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &response,
		})
		return
	}

	found := false
	for j := 0; j < int(ButtonsCount); j++ {
		var button ButtonType = ButtonType(j)
		if button.String() == requestedButton {
			found = true
		}
	}

	heldButtons = append(heldButtons, stringToButton(requestedButton))
	disabledButtons[stringToButton(requestedButton)] = true

	var response string
	if found {
		response = "Holding buttons: "
		for j := 0; j < len(heldButtons); j++ {
			response += heldButtons[j].String() + ", "
		}
		response = response[:len(response)-2]
	} else {
		response = "Button " + requestedButton + " is not a valid button. Valid buttons: "
		for j := 0; j < int(ButtonsCount); j++ {
			button := ButtonType(j)
			response += button.String() + ", "
		}
		response = response[:len(response)-2]
	}
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
}

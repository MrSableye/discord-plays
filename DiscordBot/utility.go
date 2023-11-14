package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func RSF(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func getAbsolutePath() string {
	var ret string
	for {
		fmt.Scan(&ret)
		abs := filepath.IsAbs(ret)
		exists := FileExists(ret)
		if !exists {
			fmt.Println("File does not exist. Please enter a valid path.")
		}
		if abs && exists {
			break
		}
		fmt.Println("Invalid path. Please enter an absolute path, not a relative path.")
	}
	return ret
}

func getNumber(def int) int {
	var num string
	for {
		fmt.Scan(&num)
		if num == "" {
			fmt.Println("Using default value: " + strconv.Itoa(def) + ".")
			return def
		}
		ret, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
		} else {
			return ret
		}
	}
}

func ordinal(i int) string {
	j := i % 10
	str := strconv.Itoa(i)
	if j == 1 {
		return str + "st"
	}
	if j == 2 {
		return str + "nd"
	}
	if j == 3 {
		return str + "rd"
	}
	return str + "th"
}

func printLeaderboard() string {
	var sb strings.Builder
	sort.Slice(leaderboard.Entries, func(i, j int) bool {
		return leaderboard.Entries[i].Keystrokes > leaderboard.Entries[j].Keystrokes
	})
	max, err := strconv.Atoi(S["leaderboardEntries"])
	if err != nil {
		max = 10
	}
	if len(leaderboard.Entries) < max {
		max = len(leaderboard.Entries)
	}
	for i := 0; i < max; i++ {
		sb.WriteString("" + ordinal(i+1) + ": " +
			leaderboard.Entries[i].Name + " with " +
			strconv.Itoa(leaderboard.Entries[i].Keystrokes) + " keys pressed!\n")
	}
	return sb.String()
}

func saveLeaderboard() {
	file, err := json.Marshal(leaderboard)
	check(err)
	_ = os.WriteFile("leaderboard.json", file, 0644)
}

func isAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	contains := false
	for j := 0; j < len(admins); j++ {
		if admins[j] == i.Member.User.ID {
			contains = true
			break
		}
	}
	if !contains {
		fmt.Printf("User %s tried to use an admin command\n", i.Member.User.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: SR("notAdmin", i),
			},
		})
	}
	return contains
}

func isBanned(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	for j := 0; j < len(bannedPlayers); j++ {
		if bannedPlayers[j].Id == i.Member.User.ID {
			if time.Now().Before(bannedPlayers[j].UnbanDate) {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{
							{
								Title:       "Banned",
								Description: SR("bannedMessage", i),
								Footer:      &discordgo.MessageEmbedFooter{Text: "You will be free to use this bot after " + bannedPlayers[j].UnbanDate.Format("2006-01-02")},
							},
						},
					},
				})
				return true
			} else {
				removeBanned(i.Member.User.ID)
			}
		}
	}
	days := settings.DaysConsideredTooYoung
	if i.Member.JoinedAt.After(time.Now().AddDate(0, 0, -days)) {
		unbanDate := time.Now().AddDate(0, 0, days)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "Banned",
						Description: SR("bannedMessageTooNew", i),
						Footer:      &discordgo.MessageEmbedFooter{Text: "You will be free to use this bot after " + unbanDate.Format("2006-01-02")},
					},
				},
			},
		})
		var bannedPlayer BannedPlayer = BannedPlayer{
			Id:        i.Member.User.ID,
			UnbanDate: unbanDate,
			Reason:    "Account joined server too recently",
			BannedBy:  "Bot",
			BanDate:   time.Now(),
		}
		addBanned(bannedPlayer)
		return true
	}
	return false
}

// Gets string from strings.json and replaces variables
func SR(str string, i *discordgo.InteractionCreate) string {
	var options []*discordgo.ApplicationCommandInteractionDataOption = nil
	if i.Type == discordgo.InteractionApplicationCommand || i.Type == discordgo.InteractionApplicationCommandAutocomplete {
		options = i.ApplicationCommandData().Options
	}
	ret := S[str]
	ret = strings.ReplaceAll(ret, "%NAME%", i.Member.User.Username)
	ret = strings.ReplaceAll(ret, "%ID%", i.Member.User.ID)
	ret = strings.ReplaceAll(ret, "%DATE%", time.Now().Format("2006-01-02"))
	if options != nil {
		for i := 0; i < len(options); i++ {
			ret = strings.ReplaceAll(ret, "%OPTION"+strconv.Itoa(i)+"%", options[i].StringValue())
		}
	}
	return ret
}

func addBanned(bannedPlayer BannedPlayer) bool {
	found := false
	for i := 0; i < len(bannedPlayers); i++ {
		if bannedPlayers[i].Id == bannedPlayer.Id {
			found = true
			break
		}
	}
	if found {
		return false
	}
	bannedPlayers = append(bannedPlayers, bannedPlayer)
	outJson, _ := json.Marshal(bannedPlayers)
	os.WriteFile("banned.json", outJson, 0644)
	return true
}

func getOptions(i *discordgo.InteractionCreate) map[string]string {
	if !(i.Type == discordgo.InteractionApplicationCommand || i.Type == discordgo.InteractionApplicationCommandAutocomplete) {
		return nil
	}
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]string, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt.StringValue()
	}
	return optionMap
}

func removeBanned(id string) bool {
	found := false
	for i := 0; i < len(bannedPlayers); i++ {
		if bannedPlayers[i].Id == id {
			bannedPlayers = append(bannedPlayers[:i], bannedPlayers[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return false
	}
	outJson, _ := json.Marshal(bannedPlayers)
	os.WriteFile("banned.json", outJson, 0644)
	return true
}

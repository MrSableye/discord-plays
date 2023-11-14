package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/PerformLine/go-stockutil/colorutil"
	"github.com/bwmarrin/discordgo"
)

type LeaderboardEntry struct {
	Name       string
	Id         string
	Keystrokes int
}

type Leaderboard struct {
	Entries []LeaderboardEntry
}

type BannedPlayer struct {
	Id        string
	BannedBy  string
	BanDate   time.Time
	UnbanDate time.Time
	Reason    string
}

type KeyPress struct {
	Name       string
	KeyPressed string
	Color      color.RGBA
}

var session *discordgo.Session
var lastKeyPresses []KeyPress
var bannedPlayers []BannedPlayer
var admins []string
var S map[string]string
var leaderboard Leaderboard
var keyPressCount int = 0

type ButtonType int

const (
	ButtonLeft ButtonType = iota
	ButtonRight
	ButtonUp
	ButtonDown
	ButtonA
	ButtonB
	ButtonStart
	ButtonSelect
	ButtonL
	ButtonR
	ButtonX
	ButtonY
)

var (
	commands []*discordgo.ApplicationCommand

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"screen": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandScreen(s, i)
		},
		"summary": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandSummary(s, i)
		},
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandHelp(s, i)
		},
		"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandLeaderboard(s, i)
		},
		"jail": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandJail(s, i)
		},
		"block": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandBlock(s, i)
		},
		"unblock": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandUnblock(s, i)
		},
		"dp-admin": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandAdmin(s, i)
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandStatus(s, i)
		},
		"save": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandSave(s, i)
		},
		"load": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandLoad(s, i)
		},
	}
)

var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"press_left": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonLeft)
		},
		"press_right": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonRight)
		},
		"press_up": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonUp)
		},
		"press_down": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonDown)
		},
		"press_a": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonA)
		},
		"press_b": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonB)
		},
		"press_start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonStart)
		},
		"press_select": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonSelect)
		},
		"press_l": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonL)
		},
		"press_r": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonR)
		},
		"press_x": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonX)
		},
		"press_y": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			press(s, i, ButtonY)
		},
		"hold": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			hold(s, i)
		},
	}
)

func (b *ButtonType) String() string {
	switch *b {
	case ButtonLeft:
		return "Left"
	case ButtonRight:
		return "Right"
	case ButtonUp:
		return "Up"
	case ButtonDown:
		return "Down"
	case ButtonA:
		return "A"
	case ButtonB:
		return "B"
	case ButtonStart:
		return "Start"
	case ButtonSelect:
		return "Select"
	case ButtonL:
		return "L"
	case ButtonR:
		return "R"
	case ButtonX:
		return "X"
	case ButtonY:
		return "Y"
	}
	return ""
}

func press(s *discordgo.Session, i *discordgo.InteractionCreate, button ButtonType) {
	if isBanned(s, i) {
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
	})

	truncatedName := i.Member.User.Username
	if len(truncatedName) > 7 {
		truncatedName = truncatedName[:7]
		truncatedName += "."
	}
	colorForName, _ := strconv.Atoi(i.Member.User.ID)
	r, g, b := colorutil.HslToRgb(float64(colorForName%360), 0.5, 0.5)
	lastKeyPresses = append(lastKeyPresses, KeyPress{
		Name:       truncatedName,
		KeyPressed: button.String(),
		Color:      color.RGBA{uint8(r), uint8(g), uint8(b), 255},
	})

	// Here we send
	var buf bytes.Buffer

	embeds := []*discordgo.MessageEmbed{
		{
			Image: &discordgo.MessageEmbedImage{
				URL:    "attachment://screen.gif",
				Width:  gifWidth,
				Height: gifHeight,
			},
		},
	}

	messageEdit := discordgo.NewMessageEdit(i.ChannelID, i.Message.ID)
	messageEdit.Embeds = embeds
	messageEdit.Files = []*discordgo.File{
		{
			Name:   "screen.gif",
			Reader: &buf,
		},
	}
	attachments := make([]*discordgo.MessageAttachment, 0)
	messageEdit.Attachments = &attachments
	s.ChannelMessageEditComplex(messageEdit)
	if profiling {
		log.Println("Time elapsed:", time.Since(timeStart))
	}

	// Add score to leaderboard
	if i.Member.User != nil {
		found := false
		for j := 0; j < len(leaderboard.Entries); j++ {
			if leaderboard.Entries[j].Name == i.Member.User.Username {
				leaderboard.Entries[j].Keystrokes += 1
				found = true
				break
			}
		}
		if !found {
			leaderboard.Entries = append(leaderboard.Entries, LeaderboardEntry{
				Name:       i.Member.User.Username,
				Keystrokes: 1,
			})
		}
	}
	// Save every 50 keystrokes
	keyPressCount++
	if keyPressCount >= 50 {
		keyPressCount = 0
		saveLeaderboard()
	}
}

func getButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    S["keyLText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_l",
				},
				discordgo.Button{
					Label:    S["keyUpText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_up",
				},
				discordgo.Button{
					Label:    S["keyRText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_r",
				},
				discordgo.Button{
					Label:    S["keyAText"],
					Style:    discordgo.SuccessButton,
					CustomID: "press_a",
				},
				discordgo.Button{
					Label:    S["keyXText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_x",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    S["keyLeftText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_left",
				},
				discordgo.Button{
					Label:    S["keyEmptyText"],
					Style:    discordgo.SecondaryButton,
					Disabled: true,
					CustomID: "disabled_center",
				},
				discordgo.Button{
					Label:    S["keyRightText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_right",
				},
				discordgo.Button{
					Label:    S["keyYText"],
					Style:    discordgo.SecondaryButton,
					Disabled: false,
					CustomID: "press_y",
				},
				discordgo.Button{
					Label:    S["keyBText"],
					Style:    discordgo.DangerButton,
					CustomID: "press_b",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    S["keyEmptyText"],
					Style:    discordgo.SecondaryButton,
					Disabled: true,
					CustomID: "disabled_bl",
				},
				discordgo.Button{
					Label:    S["keyDownText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_down",
				},
				discordgo.Button{
					Label:    S["keyEmptyText"],
					Style:    discordgo.SecondaryButton,
					Disabled: true,
					CustomID: "disabled_br",
				},
				discordgo.Button{
					Label:    S["keyStartText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_start",
				},
				discordgo.Button{
					Label:    S["keySelectText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_select",
				},
			},
		},
	}
}

func init() {
	stringsJson := RSF("strings.json")
	if stringsJson == "" {
		log.Fatalln("strings.json not found")
	}
	json.Unmarshal([]byte(stringsJson), &S)
	var minValue float64 = 2.0
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "screen",
			Description: S["screen"],
		},
		{
			Name:        "summary",
			Description: S["summary"],
		},
		{
			Name:        "help",
			Description: S["help"],
		},
		{
			Name:        "leaderboard",
			Description: S["leaderboard"],
		},
		{
			Name:        "jail",
			Description: S["jail"],
		},
		{
			Name:        "block",
			Description: S["block"],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user-id",
					Description: S["banOptionUserId"],
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "days",
					Description: S["banOptionDays"],
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "reason",
					Description: S["banOptionReason"],
					Required:    false,
				},
			},
		},
		{
			Name:        "unblock",
			Description: S["unblock"],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user-id",
					Description: S["unbanOptionUserId"],
					Required:    true,
				},
			},
		},
		{
			Name:        "admin",
			Description: S["admin"],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "user-id",
					Description: S["adminOptionUserId"],
					Required:    true,
				},
			},
		},
		{
			Name:        "status",
			Description: S["status"],
		},
		{
			Name:        "save",
			Description: S["save"],
		},
		{
			Name:        "load",
			Description: S["load"],
		},
		{
			Name:        "frames",
			Description: S["frames"],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "frames-pressed",
					Description: S["framesOptionFramesPressed"],
					Required:    true,
					MinValue:    &minValue,
					MaxValue:    1000,
				},
			},
		},
	}
	bannedJson := RSF("banned.json")
	if bannedJson != "" {
		json.Unmarshal([]byte(bannedJson), &bannedPlayers)
	}
	adminsJson := RSF("admins.json")
	if adminsJson != "" {
		json.Unmarshal([]byte(adminsJson), &admins)
	}
	for _, admin := range admins {
		fmt.Printf("Admin: %s\n", admin)
	}
	json.Unmarshal([]byte(RSF("leaderboard.json")), &leaderboard)
	if leaderboard.Entries == nil {
		fmt.Println("Leaderboard is nil, creating new one")
		leaderboard.Entries = make([]LeaderboardEntry, 0)
	}
}

func RunBot(BotToken string) {
	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})
	pong := get("ping")
	if pong == nil {
		log.Fatalf("Backend not running! Please start backend first!")
		return
	}
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)
	})
	session.Open()
	_, err = session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", commands)
	check(err)
	defer session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

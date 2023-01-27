package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Pokemon struct {
	Nickname string `json:"Name"`
	Name     string `json:"Type"`
	Exp      int
	Hp       int
	MaxHp    int
	Level    int
	Status   int
}

type Pokeball struct {
	Name  string
	Count int
}

type Pokeballs struct {
	Count int
	Balls []Pokeball
}

type GameData struct {
	Name  string
	Rival string
	Money int
	Johto int
	Kanto int
}

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

var bannedPlayers []BannedPlayer
var admins []string
var S map[string]string
var session *discordgo.Session
var leaderboard Leaderboard
var keyPressCount int = 0
var mutex sync.Mutex
var toggleKey int = 0
var framesSteppedPressedInit = 0

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
		"poke-jail": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandPokeJail(s, i)
		},
		"poke-ban": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandPokeBan(s, i)
		},
		"poke-unban": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandPokeUnban(s, i)
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandStatus(s, i)
		},
	}
)

func mustAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	contains := false
	for j := 0; j < len(admins); j++ {
		if admins[j] == i.Member.User.ID {
			contains = true
			break
		}
	}
	if !contains {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: SR("notAdmin", i),
			},
		})
	}
	return contains
}

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
	}
	return ""
}

func get(str string) *http.Response {
	url := "http://localhost:" + settings.Port + "/" + str
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	if settings.Debug == 1 {
		fmt.Println(url)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}

func checkOk(resp *http.Response) bool {
	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	bodystr := string(body)
	if bodystr[0] != 'o' && bodystr[1] != 'k' {
		fmt.Println("Not Ok:" + bodystr)
		return false
	}
	return true
}

func getScreen() *bytes.Reader {
	resp := get("screen")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return bytes.NewReader(body)
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
	_ = ioutil.WriteFile("leaderboard.json", file, 0644)
}

func checkBanned(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
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
	days, err := strconv.Atoi(S["guildDaysConsideredTooYoung"])
	if err != nil {
		days = 0
	}
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
	ioutil.WriteFile("banned.json", outJson, 0644)
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
	ioutil.WriteFile("banned.json", outJson, 0644)
	return true
}

func hold(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if checkBanned(s, i) {
		return
	}
	if toggleKey == 0 {
		toggleKey = 1
	} else {
		toggleKey = 0
	}
	mutex.Lock()
	checkOk(get("input?B=" + strconv.Itoa(toggleKey)))
	respondScreen(s, i)
	mutex.Unlock()
	settings.FramesSteppedPressed = framesSteppedPressedInit + toggleKey*60
}

func press(s *discordgo.Session, i *discordgo.InteractionCreate, button ButtonType) {
	if checkBanned(s, i) {
		return
	}
	mutex.Lock()
	checkOk(get("input?" + button.String() + "=1"))
	checkOk(get("step?frames=" + strconv.Itoa(settings.FramesSteppedPressed)))
	checkOk(get("input?" + button.String() + "=0"))
	checkOk(get("step?frames=" + strconv.Itoa(settings.FramesSteppedReleased)))
	respondScreen(s, i)
	mutex.Unlock()
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

func respondScreen(s *discordgo.Session, i *discordgo.InteractionCreate) {
	reader := getScreen()
	embeds := []*discordgo.MessageEmbed{
		{
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://screen.png",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: SR("footer", i) + "\nhttps://github.com/OFFTKP/pokemon-bot",
			},
		},
	}
	buttons := getButtons()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Files: []*discordgo.File{
				{Name: "screen.png", Reader: reader},
			},
			Components: buttons,
		},
	})
}

func getButtons() []discordgo.MessageComponent {
	var toggleStr string
	if toggleKey == 1 {
		toggleStr = "keyToggleOnText"
	} else {
		toggleStr = "keyToggleOffText"
	}
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
					Label:    S[toggleStr],
					Style:    discordgo.SecondaryButton,
					CustomID: "hold",
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
					Label:    S["keyEmptyText"],
					Style:    discordgo.SecondaryButton,
					Disabled: true,
					CustomID: "disabled_ll",
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
			Name:        "poke-jail",
			Description: S["poke-jail"],
		},
		{
			Name:        "poke-ban",
			Description: S["poke-ban"],
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
			Name:        "poke-unban",
			Description: S["poke-unban"],
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
			Name:        "status",
			Description: S["status"],
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
	json.Unmarshal([]byte(RSF("leaderboard.json")), &leaderboard)
	if leaderboard.Entries == nil {
		fmt.Println("Leaderboard is nil, creating new one")
		leaderboard.Entries = make([]LeaderboardEntry, 0)
	}
}

func RunBot(BotToken string) {
	framesSteppedPressedInit = settings.FramesSteppedPressed
	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
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
		fmt.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	get("step")
	session.Open()
	_, err = session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", commands)
	check(err)
	defer session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

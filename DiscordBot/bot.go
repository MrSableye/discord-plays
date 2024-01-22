package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ericpauley/go-quantize/quantize"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
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
var executablePath string
var transport *http.Transport
var profiling bool = false
var lastPressTime time.Time

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

	ButtonsCount
)

var heldButtons []ButtonType

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
		"poke-admin": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandPokeAdmin(s, i)
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
		"frames": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			commandFrames(s, i)
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

var (
	componentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"press_left": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// TODO: add way to hold dpad
			// if toggleKey == 1 {
			// 	checkOk(get("input?B=1"))
			// 	checkOk(get("step?frames=2"))
			// 	settings.FramesSteppedPressed = framesSteppedPressedInit + settings.FramesSteppedToggle*toggleKey
			// }
			press(s, i, ButtonLeft)
			// if toggleKey == 1 {
			// 	settings.FramesSteppedPressed = framesSteppedPressedInit
			// 	checkOk(get("input?B=0"))
			// }
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

func getScreen(format string) *bytes.Reader {
	resp := get("screen?format=" + format)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return bytes.NewReader(body)
}

func getScreenImageResized() image.Image {
	bytes := getScreen(settings.ImageFormat)
	var img image.Image
	var err error
	if settings.ImageFormat == "png" {
		img, err = png.Decode(bytes)
	} else if settings.ImageFormat == "jpg" {
		img, err = jpeg.Decode(bytes)
	} else if settings.ImageFormat == "bmp" {
		img, err = bmp.Decode(bytes)
	} else {
		panic("Unknown image format: " + settings.ImageFormat)
	}
	check(err)
	if settings.WidthOfImage != 0 {
		img = resize.Resize(settings.WidthOfImage, 0, img, resize.Lanczos3)
	}
	return img
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
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Toggled running: " + strconv.Itoa(toggleKey),
		},
	})
}

var gifWg sync.WaitGroup
var quantizer quantize.MedianCutQuantizer = quantize.MedianCutQuantizer{
	Aggregation: quantize.Mode,
}

func encodeAddGif(gifEncoder *gif.GIF) {
	img := getScreenImageResized()
	myPalette := quantizer.Quantize(make([]color.Color, 0, 256), img)
	palettedImg := image.NewPaletted(img.Bounds(), myPalette)
	draw.Draw(palettedImg, img.Bounds(), img, image.Point{}, draw.Src)
	gifEncoder.Image = append(gifEncoder.Image, palettedImg)
	gifEncoder.Delay = append(gifEncoder.Delay, settings.FrameDelayGif)
	gifWg.Done()
}

var timeStart time.Time
var deferredResponse bool = false

// If response is taking too much time, defer it so the discord response doesn't time out
func checkDeferResponse(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	// TODO: Defer time configurable
	if time.Since(timeStart) > time.Second*2 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
		return true
	}
	return false
}

func press(s *discordgo.Session, i *discordgo.InteractionCreate, button ButtonType) {
	if checkBanned(s, i) {
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	timeSincePress := time.Since(lastPressTime)
	lastPressTime = time.Now()
	if timeSincePress > time.Minute*2 {
		// Reset frame pressed count and held buttons
		settings.FramesSteppedPressed = initialFramesSteppedPressed
		heldButtons = make([]ButtonType, 0)
	}
	for j := 0; j < len(heldButtons); j++ {
		checkOk(get("input?" + heldButtons[j].String() + "=1"))
	}
	checkOk(get("input?" + button.String() + "=1"))
	gifEncoder := gif.GIF{}
	timeStart = time.Now()
	deferredResponse = false
	var released = false
	for j := 0; j < settings.FramesSteppedPressed+settings.FramesSteppedReleased; j += settings.FramesToSample {
		if !deferredResponse && checkDeferResponse(s, i) {
			deferredResponse = true
		}
		if !released && j >= settings.FramesSteppedPressed {
			checkOk(get("input?" + button.String() + "=0"))
			released = true
		}
		gifWg.Add(1)
		go encodeAddGif(&gifEncoder)
		checkOk(get("step?frames=" + strconv.Itoa(settings.FramesToSample)))
		gifWg.Wait()
	}
	for j := 0; j < len(heldButtons); j++ {
		checkOk(get("input?" + heldButtons[j].String() + "=0"))
	}
	gifEncoder.LoopCount = -1
	var buf bytes.Buffer
	gif.EncodeAll(&buf, &gifEncoder)
	if profiling {
		log.Println("Time elapsed:", time.Since(timeStart))
	}
	embeds := []*discordgo.MessageEmbed{
		{
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://screen.gif",
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: SR("footer", i),
			},
		},
	}
	buttons := getButtons()
	if deferredResponse {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
			Files: []*discordgo.File{
				{Name: "screen.gif", Reader: &buf},
			},
			Components: &buttons,
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: embeds,
				Files: []*discordgo.File{
					{Name: "screen.gif", Reader: &buf},
				},
				Components: buttons,
			},
		})
	}
	screenBytes := buf.Bytes()
	ioutil.WriteFile(executablePath+"/latest_save.png", screenBytes, 0644)
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

var disabledButtons [ButtonsCount]bool

func getButtons() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    S["keyLText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_l",
					Disabled: disabledButtons[ButtonL],
				},
				discordgo.Button{
					Label:    S["keyUpText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_up",
					Disabled: disabledButtons[ButtonUp],
				},
				discordgo.Button{
					Label:    S["keyRText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_r",
					Disabled: disabledButtons[ButtonR],
				},
				discordgo.Button{
					Label:    S["keyAText"],
					Style:    discordgo.SuccessButton,
					CustomID: "press_a",
					Disabled: disabledButtons[ButtonA],
				},
				discordgo.Button{
					Label:    S["keyXText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_x",
					Disabled: disabledButtons[ButtonX],
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    S["keyLeftText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_left",
					Disabled: disabledButtons[ButtonLeft],
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
					Disabled: disabledButtons[ButtonRight],
				},
				discordgo.Button{
					Label:    S["keyYText"],
					Style:    discordgo.SecondaryButton,
					CustomID: "press_y",
					Disabled: disabledButtons[ButtonY],
				},
				discordgo.Button{
					Label:    S["keyBText"],
					Style:    discordgo.DangerButton,
					CustomID: "press_b",
					Disabled: disabledButtons[ButtonB],
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
					Disabled: disabledButtons[ButtonDown],
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
					Disabled: disabledButtons[ButtonStart],
				},
				discordgo.Button{
					Label:    S["keySelectText"],
					Style:    discordgo.PrimaryButton,
					CustomID: "press_select",
					Disabled: disabledButtons[ButtonSelect],
				},
			},
		},
	}
}

func init() {
	heldButtons = make([]ButtonType, 0)
	for i := 0; i < int(ButtonsCount); i++ {
		disabledButtons[i] = false
	}
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
			Name:        "hold",
			Description: S["hold"],
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "button",
					Description: S["holdOptionButton"],
					Required:    false,
				},
			},
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
			Name:        "poke-admin",
			Description: S["poke-admin"],
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
	transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	// Enable TCP_NODELAY
	tcpConn := transport.DialContext
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := tcpConn(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return conn, nil
		}

		err = tcpConn.SetNoDelay(true)
		if err != nil {
			fmt.Println("Error setting TCP_NODELAY:", err)
		}

		return conn, nil
	}
}

func RunBot(BotToken string) {
	framesSteppedPressedInit = settings.FramesSteppedPressed
	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	executablePath = filepath.Dir(ex)
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
	get("load?path=" + executablePath + "/latest_save.png")
	get("step")
	session.Open()
	_, err = session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", commands)
	check(err)
	defer session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

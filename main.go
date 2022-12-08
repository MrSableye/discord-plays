package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"math/bits"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"

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

var session *discordgo.Session
var value int64
var processStdin io.WriteCloser
var summaryMutex sync.Mutex
var leaderboard Leaderboard

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
)

var buttonComponents []discordgo.MessageComponent = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_tl",
			},
			discordgo.Button{
				Label:    "Up",
				Style:    discordgo.PrimaryButton,
				CustomID: "press_up",
			},
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_tr",
			},
			discordgo.Button{
				Label:    "A",
				Style:    discordgo.SuccessButton,
				CustomID: "press_a",
			},
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_rr",
			},
		},
	},
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Left",
				Style:    discordgo.PrimaryButton,
				CustomID: "press_left",
			},
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_center",
			},
			discordgo.Button{
				Label:    "Right",
				Style:    discordgo.PrimaryButton,
				CustomID: "press_right",
			},
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_ll",
			},
			discordgo.Button{
				Label:    "B",
				Style:    discordgo.DangerButton,
				CustomID: "press_b",
			},
		},
	},
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_bl",
			},
			discordgo.Button{
				Label:    "Down",
				Style:    discordgo.PrimaryButton,
				CustomID: "press_down",
			},
			discordgo.Button{
				Label:    " ",
				Style:    discordgo.SecondaryButton,
				Disabled: true,
				CustomID: "disabled_br",
			},
			discordgo.Button{
				Label:    "Start",
				Style:    discordgo.PrimaryButton,
				CustomID: "press_start",
			},
			discordgo.Button{
				Label:    "Sel",
				Style:    discordgo.PrimaryButton,
				CustomID: "press_select",
			},
		},
	},
}

var (
	integerOptionMinValue = 2.0
	BotToken              = flag.String("token", RSF("token.txt"), "Bot access token")
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "screen",
			Description: "Get current screen",
		},
		{
			Name:        "party-count",
			Description: "See how many pokemon you currently have in the party",
		},
		{
			Name:        "ball-count",
			Description: "See how many pokeballs you currently have in total",
		},
		{
			Name:        "trainer",
			Description: "See general trainer description",
		},
		{
			Name:        "summary",
			Description: "Show a gif of the last few frames",
		},
		{
			Name:        "help",
			Description: "Display help dialogue",
		},
		{
			Name:        "save",
			Description: "Attemps to save the game by using a button sequence (Check image for confirmation)",
		},
		{
			Name:        "map",
			Description: "Display current map position",
		},
		{
			Name:        "leaderboard",
			Description: "Display the leaderboard",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"screen": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			sendScreenButtons(s, i)
		},
		"trainer": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ret := get("trainer")
			bs, err := ioutil.ReadAll(ret.Body)
			check(err)
			var trainer GameData
			json.Unmarshal(bs, &trainer)
			var sb strings.Builder
			sb.WriteString("Name: " + trainer.Name + "\n")
			sb.WriteString("Rival name: " + trainer.Rival + "\n")
			sb.WriteString("Money: " + strconv.Itoa(trainer.Money) + "\n")
			sb.WriteString("Johto badges: " + strconv.Itoa(bits.OnesCount8(uint8(trainer.Johto))) + "\n")
			sb.WriteString("Kanto badges: " + strconv.Itoa(bits.OnesCount8(uint8(trainer.Kanto))) + "\n")
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       "Trainer info",
							Description: sb.String(),
						},
					},
				},
			})
			check(err)
		},
		"map": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			resp := get("map")
			bs, err := ioutil.ReadAll(resp.Body)
			check(err)
			hexstr := string(bs)
			data, err := hex.DecodeString(hexstr)
			check(err)
			reader := bytes.NewReader(data)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Image: &discordgo.MessageEmbedImage{
								URL: "attachment://screen.png",
							},
							Footer: &discordgo.MessageEmbedFooter{
								Text: "https://github.com/OFFTKP/pokemon-bot",
							},
						},
					},
					Files: []*discordgo.File{
						{Name: "screen.png", Reader: reader},
					},
				},
			})
		},
		"summary": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			summaryMutex.Lock()
			defer summaryMutex.Unlock()
			get("gif")
			// bs, err := ioutil.ReadAll(resp.Body)
			// check(err)
			// hexstr := string(bs)
			// data, err := hex.DecodeString(hexstr)
			// check(err)
			data, err := ioutil.ReadFile("out.gif")
			if err != nil {
				return
			}
			reader := bytes.NewReader(data)
			embeds := []*discordgo.MessageEmbed{
				{
					Title: "GIF summary",
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://my.gif",
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "https://github.com/OFFTKP/pokemon-bot",
					},
				},
			}
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Embeds: &embeds,
				Files: []*discordgo.File{
					{Name: "my.gif", Reader: reader},
				},
			})
		},
		"party-count": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ret := get("party")
			bs, err := ioutil.ReadAll(ret.Body)
			check(err)
			var pokes []Pokemon
			json.Unmarshal(bs, &pokes)
			var sb strings.Builder
			for i, poke := range pokes {
				sb.WriteString("Pokemon " + strconv.Itoa(i+1) + ":\n")
				sb.WriteString("\tName: " + poke.Nickname + "(" + poke.Name + ")\n")
				sb.WriteString("\tLevel: " + strconv.Itoa(poke.Level) + "\n")
				sb.WriteString("\tHp: " + strconv.Itoa(poke.Hp) + "/" + strconv.Itoa(poke.MaxHp) + "\n")
				sb.WriteString("\n")
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       "You have " + strconv.Itoa(len(pokes)) + " Pokemen",
							Description: sb.String(),
						},
					},
				},
			})
			check(err)
		},
		"ball-count": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ret := get("balls")
			bs, err := ioutil.ReadAll(ret.Body)
			check(err)
			var balls Pokeballs
			json.Unmarshal(bs, &balls)
			var sb strings.Builder
			for _, ball := range balls.Balls {
				sb.WriteString(ball.Name + ": " + strconv.Itoa(ball.Count) + "\n")
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       "You have " + strconv.Itoa(balls.Count) + " balls",
							Description: sb.String(),
						},
					},
				},
			})
			check(err)
		},
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			displayHelp(s, i)
		},
		"save": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			get("save")
			saveLeaderboard()
		},
		"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ldr := printLeaderboard()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Title:       "Leaderboard",
							Description: ldr,
						},
					},
				},
			})
		},
	}
)

var (
	componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
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
	}
)

func (b *ButtonType) String() string {
	switch *b {
	case ButtonLeft:
		return "l"
	case ButtonRight:
		return "r"
	case ButtonUp:
		return "u"
	case ButtonDown:
		return "d"
	case ButtonA:
		return "a"
	case ButtonB:
		return "b"
	case ButtonStart:
		return "start"
	case ButtonSelect:
		return "select"
	}
	return ""
}

func displayHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Help",
					URL:         "https://github.com/OFFTKP/pokemon-bot",
					Type:        discordgo.EmbedTypeRich,
					Description: "Check out the github for help\n\nFeel free to contribute or raise an issue",
				},
			},
		},
	})
	check(err)
}

func respondMsg(s *discordgo.Session, i *discordgo.InteractionCreate, str string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: str,
		},
	})
	check(err)
}

func get(str string) *http.Response {
	req, _ := http.NewRequest("GET", "http://localhost:1234/"+str, nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	return resp
}

func send(str string) *http.Response {
	req, _ := http.NewRequest("GET", "http://localhost:1234/req", nil)
	q := req.URL.Query()
	q.Add("action", str)
	q.Add("val", strconv.Itoa(int(value)))
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, _ := client.Do(req)
	value = 1
	return resp
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
	max := 10
	if len(leaderboard.Entries) < 10 {
		max = len(leaderboard.Entries)
	}
	for i := 0; i < max; i++ {
		sb.WriteString("" + ordinal(i+1) + ": " +
			leaderboard.Entries[i].Name + " with " +
			strconv.Itoa(leaderboard.Entries[i].Keystrokes) + " commands sent!\n")
	}
	return sb.String()
}

func saveLeaderboard() {
	file, err := json.Marshal(&leaderboard)
	check(err)
	_ = ioutil.WriteFile("leaderboard.json", file, 0644)
}

func sendScreenButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Components: buttonComponents,
		},
	})
}

func press(s *discordgo.Session, i *discordgo.InteractionCreate, button ButtonType) {
	send(button.String())
	resp := send("screen")
	bs, err := ioutil.ReadAll(resp.Body)
	check(err)
	hexstr := string(bs)
	data, err := hex.DecodeString(hexstr)
	check(err)
	reader := bytes.NewReader(data)
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
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://screen.png",
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "https://github.com/OFFTKP/pokemon-bot\n*Last button press (" + button.String() + ") by: " + i.Member.User.Username + "*",
					},
				},
			},
			Files: []*discordgo.File{
				{Name: "screen.png", Reader: reader},
			},
			Components: buttonComponents,
		},
	})
}

func send_val(str string, val string) []byte {
	req, err := http.NewRequest("GET", "http://localhost:1234/req", nil)
	check(err)
	q := req.URL.Query()
	q.Add("action", str)
	q.Add("val", val)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	check(err)
	b, err := io.ReadAll(resp.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
	}
	return b
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func RSF(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func init() {
	flag.Parse()
	var err error
	session, err = discordgo.New("Bot " + *BotToken)
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

			if h, ok := componentsHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})
	json.Unmarshal([]byte(RSF("leaderboard.json")), &leaderboard)
	if leaderboard.Entries == nil {
		leaderboard.Entries = make([]LeaderboardEntry, 0)
	}
}

func main() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	session.Open()
	// cmds, _ := session.ApplicationCommands(session.State.User.ID, "GuildIdToDeleteCommands")
	// fmt.Printf("Old coommands size: %d\n", len(cmds))
	// for _, cmd := range cmds {
	// 	session.ApplicationCommandDelete(session.State.User.ID, "GuildIdToDeleteCommands", cmd.ID)
	// }
	_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, "", commands)
	check(err)
	defer session.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
	saveLeaderboard()
}

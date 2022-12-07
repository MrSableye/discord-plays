package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/bits"
	"net/http"
	"os"
	"os/signal"
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

var session *discordgo.Session
var value int64
var processStdin io.WriteCloser
var mutex sync.Mutex

func RSF(path string) string {
	b, err := ioutil.ReadFile(path)
	check(err)
	return string(b)
}

func init() { flag.Parse() }

func init() {
	var err error
	session, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue = 2.0
	BotToken              = flag.String("token", RSF("token.txt"), "Bot access token")
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "l",
			Description: "Hit the left button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "r",
			Description: "Hit the right button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "u",
			Description: "Hit the up button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "d",
			Description: "Hit the down button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "a",
			Description: "Hit the A button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "b",
			Description: "Hit the B button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "start",
			Description: "Hit the Start button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
		{
			Name:        "select",
			Description: "Hit the Select button",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "count",
					Description: "Count to spam button (default: 1)",
					Required:    false,
					MaxValue:    10,
				},
			},
		},
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
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"screen": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respond(s, i)
		},
		"start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("start")
			respond(s, i)
		},
		"l": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("l")
			respond(s, i)
		},
		"r": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("r")
			respond(s, i)
		},
		"u": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("u")
			respond(s, i)
		},
		"d": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("d")
			respond(s, i)
		},
		"a": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("a")
			respond(s, i)
		},
		"b": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if len(i.ApplicationCommandData().Options) == 1 {
				value = i.ApplicationCommandData().Options[0].IntValue()
			}
			send("b")
			respond(s, i)
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
				Data: &discordgo.InteractionResponseData{},
			})
			mutex.Lock()
			defer mutex.Unlock()
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
		"select": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("select")
			respond(s, i)
		},
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			displayHelp(s, i)
		},
		"save": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			get("save")
			respond(s, i)
		},
	}
)

func init() {
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp := send("screen")
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
						URL:    "attachment://screen.png",
						Width:  320,
						Height: 288,
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
	fmt.Printf("GET %s", str)
	req, _ := http.NewRequest("GET", "http://localhost:1234/"+str, nil)
	client := &http.Client{}
	resp, _ := client.Do(req)
	return resp
}

func sendGif() {

}

func send(str string) *http.Response {
	req, _ := http.NewRequest("GET", "http://localhost:1234/req", nil)
	q := req.URL.Query()
	q.Add("action", str)
	q.Add("val", strconv.Itoa(int(value)))
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)
	client := &http.Client{}
	resp, _ := client.Do(req)
	value = 1
	return resp
}

func send_val(str string, val string) []byte {
	req, err := http.NewRequest("GET", "http://localhost:1234/req", nil)
	check(err)
	q := req.URL.Query()
	q.Add("action", str)
	q.Add("val", val)
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	fmt.Println(resp)
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
}

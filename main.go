package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session
var processStdin io.WriteCloser

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
	ScreenPath            = "screen.png"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "l",
			Description: "Hit the left button",
		},
		{
			Name:        "r",
			Description: "Hit the right button",
		},
		{
			Name:        "u",
			Description: "Hit the up button",
		},
		{
			Name:        "d",
			Description: "Hit the down button",
		},
		{
			Name:        "a",
			Description: "Hit the A button",
		},
		{
			Name:        "b",
			Description: "Hit the B button",
		},
		{
			Name:        "start",
			Description: "Hit the Start button",
		},
		{
			Name:        "select",
			Description: "Hit the Select button",
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
			Name:        "spam",
			Description: "Spam a button multiple times. Dialogues go bye bye!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "button",
					Description: "Button to spam (l,r,d,u,a,b,start,select)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "spam-amount",
					Description: "Amount to press button",
					MinValue:    &integerOptionMinValue,
					MaxValue:    5,
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"screen": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("screen")
			f, err := os.Open(ScreenPath)
			check(err)
			defer f.Close()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						{
							Image: &discordgo.MessageEmbedImage{
								URL: "attachment://" + filepath.Base(ScreenPath),
							},
						},
					},
					Files: []*discordgo.File{
						{Name: filepath.Base(ScreenPath), Reader: f},
					},
				},
			})
		},
		"start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("start")
			send("screen")
			respond(s, i)
		},
		"l": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("l")
			send("screen")
			respond(s, i)
		},
		"r": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("r")
			send("screen")
			respond(s, i)
		},
		"u": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("u")
			send("screen")
			respond(s, i)
		},
		"d": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("d")
			send("screen")
			respond(s, i)
		},
		"a": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("a")
			send("screen")
			respond(s, i)
		},
		"b": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("b")
			send("screen")
			respond(s, i)
		},
		"party-count": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ret := send_val("read", "da22")
			var sb strings.Builder
			m, err := strconv.Atoi(string(ret[0]))
			if err != nil {
				return
			}
			max := uint64(m)
			for j := uint64(0); j < max; j++ {
				name := send_val("string", strconv.FormatUint(0xdb8c+(j*0xb), 16))
				sb.WriteString("Pokemon " + string('0'+j+1) + ": ")
				for _, ch := range name {
					if ch >= 0x80 && ch <= 0x99 {
						sb.WriteByte('A' + (ch - 0x80))
					} else if ch >= 0xA0 && ch <= 0xB9 {
						sb.WriteByte('a' + (ch - 0xA0))
					} else {
						sb.WriteByte(' ')
					}
				}
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You have " + string(ret[0]) + " pokemon in your party.\n" + sb.String(),
				},
			})
			check(err)
		},
		"ball-count": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ret := send_val("read", "d5fc")
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You have " + string(ret[0]) + " pokeballs.\n",
				},
			})
			check(err)
		},
		"select": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			send("select")
			send("screen")
			respond(s, i)
		},
		"spam": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respondBad(s, i)
			return
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}
			dospam := false
			spam := ""
			if option_, ok := optionMap["button"]; ok {
				option := option_.StringValue()
				if !(option == "a" || option == "b" || option == "u" || option == "d" || option == "r" ||
					option == "l" || option == "start" || option == "select") {
					respondBad(s, i)
				} else {
					dospam = true
					spam = option
				}
			}
			if option_, ok := optionMap["spam-amount"]; ok {
				option := option_.IntValue()
				if dospam {
					var j int64 = 0
					for ; j < option; j++ {
						send(spam)
					}
					send("screen")
					respond(s, i)
				}
			}
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
	f, err := os.Open(ScreenPath)
	check(err)
	defer f.Close()
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + filepath.Base(ScreenPath),
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "[Github](https://github.com/OFFTKP/pokemon-bot)",
					},
				},
			},
			Files: []*discordgo.File{
				{Name: filepath.Base(ScreenPath), Reader: f},
			},
		},
	})
	check(err)
}

func respondBad(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Bad usage of command! >:(",
		},
	})
	check(err)
}

func send(str string) {
	req, _ := http.NewRequest("GET", "http://localhost:1234/req", nil)
	q := req.URL.Query()
	q.Add("action", str)
	req.URL.RawQuery = q.Encode()
	fmt.Println(req)
	client := &http.Client{}
	resp, _ := client.Do(req)
	fmt.Println(resp)
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

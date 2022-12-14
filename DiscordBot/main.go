package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

var configFile string

type BotSettings struct {
	Token          string
	GamePath       string
	ServerPath     string
	TimeoutSeconds int
}

var webserver *exec.Cmd
var settings BotSettings

func showMenu() {
	fmt.Println("1. Run bot")
	fmt.Println("2. Configure bot")
	fmt.Println("3. Exit")
}

func configure() {
	if configFile != "" {
		fmt.Println("Config file already exists.")
		fmt.Println("Do you want to overwrite it? (y/n)")
		var input string
		for {
			fmt.Scan(&input)
			if input == "n" {
				return
			} else if input == "y" {
				break
			}
			fmt.Println("Invalid input. Please enter 'y' or 'n'.")
		}
	}
	fmt.Println("Enter bot token:")
	fmt.Println("(Not your application ID or public key!)")
	fmt.Println("(https://discord.com/developers/applications -> Bot -> Copy Token)")
	var token string
	fmt.Scan(&token)
	fmt.Println("Enter absolute path to GameboyWebserver executable:")
	serverPath := GetAbsolutePath()
	fmt.Println("Enter absolute path to game ROM (works with Gold/Silver but not Crystal):")
	gamePath := GetAbsolutePath()
	settings = BotSettings{token, gamePath, serverPath, 5}
	settingsJson, err := json.Marshal(settings)
	check(err)
	fmt.Println("Writing config file...")
	ioutil.WriteFile("config.json", settingsJson, 0644)
	configFile = RSF("config.json")
	fmt.Println("Bot configured successfully.")
}

func startServer() bool {
	webserver = exec.Command(settings.ServerPath, settings.GamePath)
	err := webserver.Start()
	check(err)
	// Check if server started successfully
	fmt.Println("Starting GameboyWebserver on port 1234...")
	i := 0
	var resp *http.Response
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("Pinging server...")
		resp = get("ping")
		if resp != nil {
			break
		}
		i++
		if i > settings.TimeoutSeconds {
			fmt.Println("Server failed to start.")
			fmt.Println("You can increase the timeout in config.json.")
			return false
		}
	}
	bs, err := ioutil.ReadAll(resp.Body)
	check(err)
	respStr := string(bs)
	if respStr != "pong" {
		fmt.Println("Server failed to start.")
		return false
	} else {
		fmt.Println("Server started successfully.")
	}
	return true
}

func showPanel() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Bot is running. Press Ctrl+C to stop.")
	fmt.Println("Control panel: http://localhost:1234/panel.html (TODO)")
	<-stop
}

func run() {
	if configFile == "" {
		fmt.Println("Bot is not configured. Please configure it first.")
		return
	}
	fmt.Println("Starting webserver...")
	gwStarted := startServer()
	if !gwStarted {
		return
	}
	fmt.Println("Starting Discord bot...")
	go RunBot(settings.Token)
	showPanel()
}

func exit() {
	if webserver != nil {
		webserver.Process.Kill()
	}
	os.Exit(0)
}

func main() {
	configFile = RSF("config.json")
	json.Unmarshal([]byte(configFile), &settings)
	for {
		showMenu()
		var input int
		fmt.Scan(&input)
		switch input {
		case 1:
			run()
		case 2:
			configure()
		case 3:
			exit()
		}
		fmt.Println("Going back to menu:")
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

var configFile string

type BotSettings struct {
	Token                  string
	GamePath               string
	ServerPath             string
	TimeoutSeconds         int
	Port                   string
	StartCommand           string
	FramesSteppedPressed   int
	FramesSteppedReleased  int
	FramesSteppedToggle    int
	FramesToSample         int
	WidthOfImage           uint
	ImageFormat            string
	Debug                  int
	FrameDelayGif          int
	DaysConsideredTooYoung int
}

func DefaultBotSettings() BotSettings {
	var newSettings BotSettings
	newSettings.FramesSteppedPressed = 10
	newSettings.FramesSteppedReleased = 50
	newSettings.FramesSteppedToggle = 20
	newSettings.FramesToSample = 5
	newSettings.WidthOfImage = 0
	newSettings.ImageFormat = "bmp"
	newSettings.Debug = 0
	newSettings.FrameDelayGif = 10
	newSettings.DaysConsideredTooYoung = 0
	return newSettings
}

var webserver *exec.Cmd
var settings BotSettings
var initialFramesSteppedPressed int

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
	fmt.Println("Enter absolute path to backend executable (example: C:/Users/JohnDoe/Desktop/SkyEmu.exe):")
	serverPath := GetAbsolutePath()
	fmt.Println("Enter absolute path to game ROM (example: C:/Users/JohnDoe/Desktop/pokemon_gold.gba):")
	gamePath := GetAbsolutePath()
	timeout := 5
	for {
		fmt.Println("Enter timeout in seconds for server to start (default: 5):")
		timeout = GetNumber(timeout)
		if timeout > 0 {
			break
		}
		fmt.Println("Timeout must be greater than 0")
	}
	port := 1234
	for {
		fmt.Println("Enter port number for webserver (default: 1234):")
		port = GetNumber(port)
		if port > 0 && port < 65536 {
			break
		}
		fmt.Println("Port must be between 1 and 65535")
	}
	fmt.Println("Which backend do you want to use?")
	fmt.Println("1. SkyEmu")
	fmt.Println("2. Other...")
	var backend int
	var startCommand string
	for {
		backend = GetNumber(1)
		if backend > 0 && backend < 4 {
			switch backend {
			case 1:
				startCommand = "%SERVERPATH% http_server %PORT% %GAMEPATH%"
			case 2:
				fmt.Println("Enter the command that will start the webserver:")
				fmt.Println("Use %SERVERPATH% for the path to the executable")
				fmt.Println("Use %GAMEPATH% for the path to the game ROM")
				fmt.Println("Use %PORT% for the port number")
				fmt.Scan(&startCommand)
			}
			break
		}
		fmt.Println("Invalid input. Please enter a number between 1 and 3.")
	}
	settings = BotSettings{token, gamePath, serverPath, timeout, strconv.Itoa(port), startCommand, 5, 60, 30, 5, 0, "bmp", 0, 10, 0}
	settingsJson, err := json.Marshal(settings)
	check(err)
	fmt.Println("Writing config file...")
	os.WriteFile("config.json", settingsJson, 0644)
	configFile = RSF("config.json")
	fmt.Println("Bot configured successfully.")
}

func getWebserverCommand() *exec.Cmd {
	command := settings.StartCommand
	command = strings.ReplaceAll(command, "%SERVERPATH%", settings.ServerPath)
	command = strings.ReplaceAll(command, "%GAMEPATH%", settings.GamePath)
	command = strings.ReplaceAll(command, "%PORT%", settings.Port)
	fmt.Println("Starting webserver with command: " + command)
	split := strings.Split(command, " ")
	return exec.Command(split[0], split[1:]...)
}

func startServer() bool {
	webserver = getWebserverCommand()
	err := webserver.Start()
	check(err)
	// Check if server started successfully
	fmt.Println("Starting backend on port " + settings.Port + "...")
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
	return true
}

func showPanel() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Bot is running. Press Ctrl+C to stop.")
	fmt.Println("Control panel: http://localhost:4321/panel.html (Not yet implemented) (TODO)")
	// TODO: Implement control panel, on Go side for easier execution of admin commands
	<-stop
}

func run() {
	if configFile == "" {
		fmt.Println("Bot is not configured. Please configure it first.")
		return
	}
	pong := get("ping")
	if pong == nil {
		gwStarted := startServer()
		if !gwStarted {
			return
		}
	} else {
		fmt.Println("Server already running. Skipping startup.")
	}
	fmt.Println("Starting Discord bot...")
	go RunBot(settings.Token)
	showPanel()
}

func main() {
	settings = DefaultBotSettings()
	configFile = RSF("config.json")
	if configFile != "" {
		err := json.Unmarshal([]byte(configFile), &settings)
		check(err)
		initialFramesSteppedPressed = settings.FramesSteppedPressed
		run()
	} else {
		configure()
	}
}

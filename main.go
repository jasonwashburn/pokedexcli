package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	next     string
	previous string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

var supportedCommands map[string]cliCommand

func initCommands() {
	supportedCommands = map[string]cliCommand{
		"map": {
			name:        "map",
			description: "Display the next 20 locations in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the previous 20 locations in the Pokemon world",
			callback:    commandMapB,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}

func main() {
	config := Config{}
	initCommands()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cleanInput := cleanInput(input)

		command, ok := supportedCommands[cleanInput[0]]
		if !ok {
			fmt.Println("Unknown command")
		} else {
			if err := command.callback(&config); err != nil {
				fmt.Println(err)
			}
		}

	}
}

func cleanInput(text string) []string {
	loweredStrings := strings.ToLower(text)
	return strings.Fields(loweredStrings)
}

func commandHelp(config *Config) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	for _, cmd := range supportedCommands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandExit(config *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(config *Config) error {
	var url string
	if config.next == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = config.next
	}
	locationAreaResponse, err := getLocationArea(url)
	if err != nil {
		return err
	}

	for _, area := range locationAreaResponse.Results {
		fmt.Println(area.Name)
	}

	config.previous = locationAreaResponse.Previous
	config.next = locationAreaResponse.Next

	return nil
}

func commandMapB(config *Config) error {
	url := config.previous
	if url == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	locationAreaResponse, err := getLocationArea(url)
	if err != nil {
		return err
	}

	for _, area := range locationAreaResponse.Results {
		fmt.Println(area.Name)
	}

	config.previous = locationAreaResponse.Previous
	config.next = locationAreaResponse.Next

	return nil
}

func getLocationArea(url string) (LocationAreaResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("failed to read response body: %v", err)
	}

	locationAreaResponse := LocationAreaResponse{}
	if err := json.Unmarshal(body, &locationAreaResponse); err != nil {
		return LocationAreaResponse{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return locationAreaResponse, nil
}

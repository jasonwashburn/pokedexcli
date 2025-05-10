package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jasonwashburn/pokedexcli/internal/pokecache"
)

type Config struct {
	next     string
	previous string
	cache    *pokecache.Cache
	pokedex  map[string]Pokemon
}

func initConfig() Config {
	return Config{
		pokedex: make(map[string]Pokemon),
	}
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, ...string) error
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
		"explore": {
			name:        "explore",
			description: "Explore a location in the Pokemon world",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Display your Pokedex",
			callback:    commandPokedex,
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
	config := initConfig()
	config.cache = pokecache.NewCache(5 * time.Second)
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
			if err := command.callback(&config, cleanInput[1:]...); err != nil {
				fmt.Println(err)
			}
		}

	}
}

func cleanInput(text string) []string {
	loweredStrings := strings.ToLower(text)
	return strings.Fields(loweredStrings)
}

func commandHelp(config *Config, args ...string) error {
	fmt.Print("Welcome to the Pokedex!\nUsage:\n\n")
	for _, cmd := range supportedCommands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandExit(config *Config, args ...string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(config *Config, args ...string) error {
	var url string
	if config.next == "" {
		url = "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	} else {
		url = config.next
	}
	locationAreaResponse, err := getLocationArea(config, url)
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

func commandMapB(config *Config, args ...string) error {
	url := config.previous
	if url == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	locationAreaResponse, err := getLocationArea(config, url)
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

func cachedRequest(config *Config, url string) (body []byte, err error) {
	if cached, ok := config.cache.Get(url); ok {
		fmt.Println("using cached data")
		body = cached
	} else {
		fmt.Println("fetching data from:", url)
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch data: %v", err)
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
		config.cache.Add(url, body)
	}

	return body, nil
}

func getLocationArea(config *Config, url string) (LocationAreaListResponse, error) {
	var body []byte
	body, err := cachedRequest(config, url)
	if err != nil {
		return LocationAreaListResponse{}, err
	}
	locationAreaResponse := LocationAreaListResponse{}
	if err := json.Unmarshal(body, &locationAreaResponse); err != nil {
		return LocationAreaListResponse{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return locationAreaResponse, nil
}

func commandExplore(config *Config, args ...string) error {
	location := args[0]
	if location == "" {
		return fmt.Errorf("no location provided")
	}
	if err := getPokemonByLocationArea(config, location); err != nil {
		return err
	}
	return nil
}

func getPokemonByLocationArea(config *Config, location string) error {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", location)
	body, err := cachedRequest(config, url)
	if err != nil {
		return err
	}
	locationAreaResponse := LocationAreaResponse{}
	if err := json.Unmarshal(body, &locationAreaResponse); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	fmt.Printf("Exploring %s...\nFound Pokemon:\n", location)
	for _, encounter := range locationAreaResponse.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(config *Config, args ...string) error {
	pokemonName := args[0]
	if pokemonName == "" {
		return fmt.Errorf("no pokemon name provided")
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)
	pokemon, err := getPokemonByName(config, pokemonName)
	if err != nil {
		return err
	}
	caught := tryCatchPokemon(config, pokemon)
	if caught {
		fmt.Printf("%s was caught!\n", pokemonName)
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}
	return nil
}

func commandInspect(config *Config, args ...string) error {
	pokemonName := args[0]
	if pokemonName == "" {
		return fmt.Errorf("no pokemon name provided")
	}
	pokemon, ok := config.pokedex[pokemonName]
	if !ok {
		return fmt.Errorf("pokemon not found in pokedex")
	}
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Printf("Stats:\n")
	for stat, value := range pokemon.GetStats() {
		fmt.Printf("  -%s: %d\n", stat, value)
	}
	fmt.Printf("Types:\n")
	for _, ptype := range pokemon.GetTypeNames() {
		fmt.Printf("  -%s\n", ptype)
	}
	return nil
}

func commandPokedex(config *Config, _ ...string) error {
	//Your Pokedex:
	//  - pidgey
	//  - caterpie
	fmt.Println("Your Pokedex:")
	for pokemon := range config.pokedex {
		fmt.Printf(" - %s\n", pokemon)
	}
	return nil
}

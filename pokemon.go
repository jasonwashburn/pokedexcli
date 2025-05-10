package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type EncounteredPokemon struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PokemonEncounter struct {
	Pokemon EncounteredPokemon `json:"pokemon"`
}

type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Types          []Type `json:"types"`
	Stats          []Stat `json:"stats"`
}

type Stat struct {
	Stat     StatMetadata `json:"stat"`
	BaseStat int          `json:"base_stat"`
}

type StatMetadata struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (p Pokemon) GetStats() map[string]int {
	stats := make(map[string]int)
	for _, stat := range p.Stats {
		stats[stat.Stat.Name] = stat.BaseStat
	}
	return stats
}

func (p Pokemon) GetTypeNames() []string {
	names := make([]string, len(p.Types))
	for i, ptype := range p.Types {
		names[i] = ptype.Type.Name
	}
	return names
}

type Ability struct {
	Ability AbilityMetadata `json:"ability"`
}

type AbilityMetadata struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Type struct {
	Type TypeMetadata `json:"type"`
}

type TypeMetadata struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func getPokemonByName(config *Config, name string) (Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)

	body, err := cachedRequest(config, url)
	if err != nil {
		return Pokemon{}, err
	}

	var pokemon Pokemon
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return Pokemon{}, err
	}

	return pokemon, nil
}

func tryCatchPokemon(config *Config, pokemon Pokemon) bool {
	random := rand.Intn(100)
	threshold := int(float64(600-pokemon.BaseExperience) / 600.0 * 100.0)
	if random >= threshold {
		config.pokedex[pokemon.Name] = pokemon
		return true
	}
	return false
}

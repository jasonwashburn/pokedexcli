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

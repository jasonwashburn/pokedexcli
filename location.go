package main

type LocationArea struct {
	Name string
	Url  string
}

type LocationAreaListResponse struct {
	Count    int
	Next     string
	Previous string
	Results  []LocationArea
}

type LocationAreaResponse struct {
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

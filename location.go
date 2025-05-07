package main

type LocationArea struct {
	Name string
	Url  string
}

type LocationAreaResponse struct {
	Count    int
	Next     string
	Previous string
	Results  []LocationArea
}

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sam-brook/pokemoncli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(string) error
}

var commands map[string]cliCommand

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	result := strings.Fields(lowercase)
	return result
}

func commandExit(argument string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(argument string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

var current_location location_area
var cache pokecache.Cache

type location_area struct {
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
	} `json:"results"`
}

func commandMap(argument string) error {
	var url string
	if current_page == 0 {
		url = "https://pokeapi.co/api/v2/location-area/"
	} else {
		url = current_location.Next
	}

	val, exists := cache.Get(url)
	var next_location location_area

	if exists {
		err := json.Unmarshal(val, &next_location)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode > 299 {
			failedErr := fmt.Sprintf("Response failed with status code %d", res.StatusCode)
			return errors.New(failedErr)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &next_location)
		if err != nil {
			return err
		}
		cache.Add(url, body)
	}

	for _, loc := range next_location.Results {
		fmt.Println(loc.Name)
	}
	current_location = next_location
	current_page++

	return nil
}

func commandMapB(argument string) error {

	if current_page <= 1 {
		fmt.Println("you're on the first page")
		return nil
	}

	url := current_location.Previous
	val, exists := cache.Get(url)
	var next_location location_area

	if exists {
		err := json.Unmarshal(val, &next_location)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode > 299 {
			failedErr := fmt.Sprintf("Response failed with status code %d", res.StatusCode)
			return errors.New(failedErr)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &next_location)
		if err != nil {
			return err
		}
		for _, location := range next_location.Results {
			fmt.Println(location.Name)
		}
		cache.Add(url, body)
	}

	current_location = next_location
	current_page--

	return nil
}

type Type_Info struct {
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

type Pokemon struct {
	Name       string      `json:"name"`
	Experience int         `json:"base_experience"`
	Height     int         `json:"height"`
	Weight     int         `json:"weight"`
	Stats      []Stat      `json:"stats"`
	Types      []Type_Info `json:"types"`
}

type Response struct {
	PokemonEncounters []Pokemon `json:"pokemon_encounters"`
}

func commandExplore(argument string) error {
	url := "https://pokeapi.co/api/v2/location-area/" + argument

	val, exists := cache.Get(url)
	if exists {
		var location_response Response
		err := json.Unmarshal(val, &location_response)
		if err != nil {
			return err
		}
		for _, encounter := range location_response.PokemonEncounters {
			name := encounter.Name
			fmt.Println(name)
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode > 299 {
			failedErr := fmt.Sprintf("Response failed with status code %d", res.StatusCode)
			return errors.New(failedErr)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var location_response Response
		err = json.Unmarshal(body, &location_response)
		if err != nil {
			return err
		}
		for _, encounter := range location_response.PokemonEncounters {
			name := encounter.Name
			fmt.Println(name)
		}
		cache.Add(url, body)
	}
	return nil
}

type Stat struct {
	Base_stat int `json:"base_stat"`
	Stat_info struct {
		Name string `json:"name"`
	} `json:"stat"`
}

func commandInspect(argument string) error {
	pokemon, ok := pokedex[argument]
	if !ok {
		fmt.Println("you have not caught that pokemon")
	} else {
		fmt.Printf("Name: %v\nHeight: %v\nWeight: %v\nStats:\n", pokemon.Name, pokemon.Height, pokemon.Weight)
		for _, stats := range pokemon.Stats {
			fmt.Printf("  -%v: %v\n", stats.Stat_info.Name, stats.Base_stat)
		}
		fmt.Println("Types:")
		for _, t := range pokemon.Types {
			fmt.Println(" - " + t.Type.Name)
		}
	}
	return nil
}

func commandCatch(argument string) error {
	url := "https://pokeapi.co/api/v2/pokemon/" + argument
	var pokemon Pokemon

	val, exists := cache.Get(url)
	if exists {
		err := json.Unmarshal(val, &pokemon)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(url)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode > 299 {
			failedErr := fmt.Sprintf("Response failed with status code %d, maybe the pokemon name was invalid", res.StatusCode)
			return errors.New(failedErr)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &pokemon)
		if err != nil {
			return err
		}

		cache.Add(url, body)
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	if catch_pokemon(pokemon) {
		fmt.Println(pokemon.Name + " was caught!")
		pokedex[argument] = pokemon
	} else {
		fmt.Println(pokemon.Name + " escaped")
	}

	return nil
}

func catch_pokemon(pokemon Pokemon) bool {
	num := rand.Intn(pokemon.Experience)

	if num <= 50 {
		return true
	} else {
		return false
	}
}

var pokedex map[string]Pokemon

func commandPokedex(argument string) error {
	fmt.Println("Your Pokedex: ")
	if len(pokedex) == 0 {
		return errors.New("No pokemon in pokedex")
	}
	for _, pokemon := range pokedex {
		fmt.Println(" - " + pokemon.Name)
	}
	return nil
}

var current_page int

func main() {
	cache = pokecache.NewCache(2 * time.Second)

	current_page = 0

	pokedex = map[string]Pokemon{}
	commands = map[string]cliCommand{
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
		"map": {
			name:        "map",
			description: "Displays the next locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous locations",
			callback:    commandMapB,
		},
		"explore": {
			name:        "explore",
			description: "Displays the pokemon in a given location",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempts to catch the given pokemon",
			callback:    commandCatch,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays all the pokemon you have caught",
			callback:    commandPokedex,
		},
		"inspect": {
			name:        "inspect",
			description: "Displays information on the given pokemon if it is in your pokedex",
			callback:    commandInspect,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		text := scanner.Text()
		cleanText := cleanInput(text)
		command, ok := commands[cleanText[0]]
		if ok {
			if len(cleanText) > 1 {
				err := command.callback(cleanText[1])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				err := command.callback("")
				if err != nil {
					fmt.Println(err)
				}
			}
		} else {
			fmt.Println("Unknown command")
		}
		fmt.Print("Pokedex > ")
	}
}

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
	"strconv"
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

var current_location_id int
var cache pokecache.Cache

type location_area struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// NOTE for future self, if possible, make the requests and store them in a slice, then when going to a new location make more requests, but if the
// current slice size is greater than the location requested, just loop through the slice and print the values there
func commandMap(argument string) error {
	for i := 0; i < 20; i++ {
		url := "https://pokeapi.co/api/v2/location-area/" + strconv.Itoa(i+current_location_id)
		val, exists := cache.Get(url)
		if exists {
			var currentLoc location_area
			err := json.Unmarshal(val, &currentLoc)
			if err != nil {
				return err
			}
			fmt.Println(currentLoc.Name)
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

			var currentLoc location_area
			err = json.Unmarshal(body, &currentLoc)
			if err != nil {
				return err
			}
			fmt.Println(currentLoc.Name)
			cache.Add(url, body)
		}

	}

	current_location_id += 20
	return nil
}

func commandMapB(argument string) error {
	if current_location_id > 20 {
		current_location_id -= 20
	}

	if current_location_id < 20 {
		fmt.Println("you're on the first page")
		return nil
	}

	for i := 0; i < 20; i++ {
		url := "https://pokeapi.co/api/v2/location-area/" + strconv.Itoa(i+current_location_id)
		val, exists := cache.Get(url)
		if exists {
			var currentLoc location_area
			err := json.Unmarshal(val, &currentLoc)
			if err != nil {
				return err
			}
			fmt.Println(currentLoc.Name)
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

			var currentLoc location_area
			err = json.Unmarshal(body, &currentLoc)
			if err != nil {
				return err
			}
			fmt.Println(currentLoc.Name)
			cache.Add(url, body)
		}
	}
	return nil
}

type Response struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
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
			name := encounter.Pokemon.Name
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
			name := encounter.Pokemon.Name
			fmt.Println(name)
		}
		cache.Add(url, body)
	}
	return nil
}

type Pokemon struct {
	Name       string `json:"name"`
	Experience int    `json:"base_experience"`
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

func main() {
	cache = pokecache.NewCache(2 * time.Second)

	current_location_id = 1

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

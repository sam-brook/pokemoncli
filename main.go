package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var commands map[string]cliCommand

func cleanInput(text string) string {
	lowercase := strings.ToLower(text)
	result := strings.Fields(lowercase)
	return result[0]
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

var current_location_id int
var seen_locations []string

type location_area struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// NOTE for future self, if possible, make the requests and store them in a slice, then when going to a new location make more requests, but if the
// current slice size is greater than the location requested, just loop through the slice and print the values there
func commandMap() error {
	if current_location_id >= len(seen_locations)+1 {
		for i := 0; i < 20; i++ {
			url := "https://pokeapi.co/api/v2/location-area/" + strconv.Itoa(i+current_location_id)

			res, err := http.Get(url)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}

			if res.StatusCode > 299 {
				failedErr := fmt.Sprintf("Response failed with status code %d", res.StatusCode)
				return errors.New(failedErr)
			}
			var currentLoc location_area
			err = json.Unmarshal(body, &currentLoc)
			if err != nil {
				return err
			}
			seen_locations = append(seen_locations, currentLoc.Name)
		}
	}

	for i := current_location_id; i < current_location_id+20; i++ {
		fmt.Println(seen_locations[i-1])
	}
	current_location_id += 20
	return nil
}

func commandMapB() error {
	if current_location_id < 20 {
		fmt.Println("you're on the first page")
	} else {
		current_location_id -= 20
		for i := 0; i < 20; i++ {
			fmt.Println(seen_locations[i+current_location_id-1])
		}
	}
	return nil
}

func main() {
	current_location_id = 1
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
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		text := scanner.Text()
		cleanText := cleanInput(text)
		command, ok := commands[cleanText]
		if ok {
			err := command.callback()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Unknown command")
		}
		fmt.Print("Pokedex > ")
	}
}

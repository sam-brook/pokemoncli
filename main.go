package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	result := strings.Fields(lowercase)
	return result
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		text := scanner.Text()
		cleanText := cleanInput(text)
		fmt.Printf("Your command was: %s\n", cleanText[0])
		fmt.Print("Pokedex > ")
	}
}

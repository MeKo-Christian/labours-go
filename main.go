package main

import (
	"log"

	"labours-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

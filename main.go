package main

import (
	"labours-go/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

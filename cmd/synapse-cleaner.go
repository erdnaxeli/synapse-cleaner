package main

import (
	"fmt"
	"os"

	synapsecleaner "github.com/erdnaxeli/synapse-cleaner"
)

type Room struct {
	CanonicalAlias string
	Id             string
	Name           string
}

func main() {
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Print("Please set the ACCESS_TOKEN env var.")
		os.Exit(1)
	}

	if len(os.Args) != 6 || os.Args[1] != "purge-rooms" || os.Args[2] != "--user" || os.Args[4] != "--server" {
		fmt.Printf("Usage: %s purge-rooms --user <user> --server <server>", os.Args[0])
		os.Exit(1)
	}

	user := os.Args[3]
	server := os.Args[5]

	purger := synapsecleaner.NewRoomsPurger(accessToken, user, server)
	err := purger.Do()
	if err != nil {
		fmt.Print("Error: ", err)
		os.Exit(1)
	}

	os.Exit(0)
}

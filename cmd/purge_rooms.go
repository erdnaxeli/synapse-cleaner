package main

import (
	"fmt"
	"os"

	synapsecleaner "github.com/erdnaxeli/synapse-cleaner"
)

type ApiContext struct {
	AccessToken string `required:"" env:"ACCESS_TOKEN"`
	Server      string `required:"" short:"s"`
}

type PurgeRoomsCmd struct {
	ApiContext

	User string `arg:""`
}

func (cmd *PurgeRoomsCmd) Run() error {
	purger := synapsecleaner.NewRoomsPurger(cmd.AccessToken, cmd.User, cmd.ApiContext.Server)
	err := purger.Do()
	if err != nil {
		fmt.Print("Error: ", err)
		os.Exit(1)
	}
	return nil
}

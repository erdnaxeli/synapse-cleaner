package main

import (
	synapsecleaner "github.com/erdnaxeli/synapse-cleaner"
)

type DatabaseContext struct {
	DatabaseUri string `required:"" env:"SYNAPSE_DB_URI" short:"d"`
}

type PurgeMediasCmd struct {
	DatabaseContext

	MediaDirectory string `required:"" short:"m"`
}

func (cmd *PurgeMediasCmd) Run() error {
	purger := &synapsecleaner.MediasPurger{
		DatabaseUri:     cmd.DatabaseUri,
		MediasDirectory: cmd.MediaDirectory,
	}
	err := purger.Run()

	return err
}

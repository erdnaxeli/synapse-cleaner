package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	PurgeMedias PurgeMediasCmd `cmd:""`
	PurgeRooms  PurgeRoomsCmd  `cmd:""`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
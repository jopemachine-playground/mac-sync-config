package main

import (
	"log"
	"os"
	"sort"

	API "github.com/jopemachine/mac-sync/src"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mac-sync",
		Usage: "Sync configs and programs between macs or accounts.",
		Commands: []*cli.Command{
			{
				Name:    "sync",
				Aliases: []string{"s"},
				Usage:   "Sync configs",
				Action: func(*cli.Context) error {
					API.SyncPrograms()
					return nil
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Edit config",
				Action: func(*cli.Context) error {
					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

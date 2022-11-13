package main

import (
	"log"
	"os"

	API "github.com/jopemachine/mac-sync/src"
	"github.com/urfave/cli/v2"
)

func main() {
	if API.IsRootUser() {
		API.Logger.Error("Running mac-sync as root is not allowed.\nIf you want to install some programs as root, prepend 'sudo' into the install command.")
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "mac-sync",
		Usage: "Sync the config files and programs between macs through Github.",
		Commands: []*cli.Command{
			{
				Name:  "push",
				Usage: "Push the local config files to the remote repository",
				Action: func(*cli.Context) error {
					API.PushConfigFiles()
					return nil
				},
			},
			{
				Name:      "pull",
				Usage:     "Pull the config files from the remote repository",
				ArgsUsage: "Filter to basename of the config file",
				Action: func(c *cli.Context) error {
					API.PullRemoteConfigs(c.Args().First())
					return nil
				},
			},
			{
				Name:  "sync",
				Usage: "Sync programs with the remote repository",
				Action: func(*cli.Context) error {
					API.SyncPrograms()
					return nil
				},
			},
			{
				Name:    "clear-cache",
				Aliases: []string{"c"},
				Usage:   "Clear cache",
				Action: func(*cli.Context) error {
					API.ClearCache()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

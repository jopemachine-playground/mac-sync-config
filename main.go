package main

import (
	"log"
	"os"

	API "github.com/jopemachine/mac-sync-config/src"
	API_UTILS "github.com/jopemachine/mac-sync-config/src/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	if API_UTILS.IsRootUser() {
		API.Logger.Error("Running mac-sync-config as root is not allowed.\nIf you want to install some programs as root, prepend 'sudo' into the install command.")
		os.Exit(1)
	}

	app := &cli.App{
		Name:  "mac-sync-config",
		Usage: "Sync the config files between macs through Github",
		Commands: []*cli.Command{
			{
				Name:  "push",
				Usage: "Push the local config files to the remote repository",
				Action: func(*cli.Context) error {
					API.PushConfigFiles()
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "overwrite",
						Aliases:     []string{"o"},
						Destination: &API.Flag_OverWrite,
					},
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
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "overwrite",
						Aliases:     []string{"o"},
						Destination: &API.Flag_OverWrite,
					},
				},
			},
			{
				Name:  "list",
				Usage: "Show the configuration files list",
				Action: func(*cli.Context) error {
					API.PrintConfig()
					return nil
				},
			},
			{
				Name:  "clear-cache",
				Usage: "Clear cache used in \"pull\" command",
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

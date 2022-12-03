package main

import (
	"log"
	"os"

	API "github.com/jopemachine/mac-sync-config/src"
	Commands "github.com/jopemachine/mac-sync-config/src/commands"
	Utils "github.com/jopemachine/mac-sync-config/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	if Utils.IsRootUser() {
		API.Logger.Error("Running mac-sync-config as root is not allowed.\nIf you want to install some programs as root, prepend 'sudo' into the install command.")
		os.Exit(1)
	}

	app := &cli.App{
		Name:      "mac-sync-config",
		Usage:     "Sync your config files between macs through your Github repository.",
		UsageText: "mac-sync-config command [command options] [arguments...]",
		Version:   "0.1.0",
		Suggest:   true,
		Commands: []*cli.Command{
			{
				Name:  "push",
				Usage: "Push the local config files to the remote repository",
				Action: func(*cli.Context) error {
					Commands.PushConfigFiles()
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
					Commands.PullRemoteConfigs(c.Args().First())
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
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "Show the configuration files list",
				Action: func(*cli.Context) error {
					Commands.PrintMacSyncConfigs()
					return nil
				},
			},
			{
				Name:    "switch-profile",
				Aliases: []string{"profile"},
				Usage:   "Switch the profile. This could be useful when you need to the configuration set",
				Action: func(c *cli.Context) error {
					Commands.SwitchProfile(c.Args().First())
					return nil
				},
			},
			// TODO: Remove this command in release build.
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

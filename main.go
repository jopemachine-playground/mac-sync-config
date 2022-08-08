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
		Usage: "Sync config files and programs between macs or accounts using Git.",
		Commands: []*cli.Command{
			{
				Name:    "upload-configs",
				Aliases: []string{"u"},
				Usage:   "Upload local configs",
				Action: func(*cli.Context) error {
					API.UploadConfigs()
					return nil
				},
			},
			{
				Name:    "download-configs",
				Aliases: []string{"d"},
				Usage:   "Download remote configs",
				Action: func(*cli.Context) error {
					API.DownloadConfigs()
					return nil
				},
			},
			{
				Name:    "sync-programs",
				Aliases: []string{"s"},
				Usage:   "Sync programs with remote",
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

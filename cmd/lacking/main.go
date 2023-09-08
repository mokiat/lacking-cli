package main

import (
	"log"
	"os"

	"github.com/mokiat/lacking-cli/internal/distribution/linux"
	"github.com/mokiat/lacking-cli/internal/distribution/macos"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name: "lacking",
		Action: func(*cli.Context) error {
			return nil
		},
		Commands: []cli.Command{
			{
				Name:  "dist",
				Usage: "Builds an application distribution",
				Subcommands: cli.Commands{
					{
						Name:   "linux",
						Usage:  "Builds a Linux distribution",
						Action: linux.Package,
					},
					{
						Name:   "macos",
						Usage:  "Builds a MacOS distribution",
						Action: macos.Package,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	Version   string
	projectID string
	secretID  string
)

func main() {
	app := &cli.App{
		Name:      "Espresso Keystore and Secrets Manager",
		Usage:     "Update a Secret Manager secret with Sequencer private keys and DB keys.",
		Version:   Version,
		Copyright: "(c) 2024 Nethermind",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project-id",
				Usage:       "Google Cloud Project ID",
				Destination: &projectID,
				EnvVars:     []string{"PROJECT_ID"},
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "secret-id",
				Usage:       "Secret Manager secret ID",
				Destination: &secretID,
				EnvVars:     []string{"SECRET_ID"},
				Required:    true,
			},
		},
		Commands: []*cli.Command{
			keystoreCMD(),
			dbKeysCMD(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

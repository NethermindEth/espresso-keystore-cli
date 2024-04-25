package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	version   = "0.2.0"
	projectID = os.Getenv("PROJECT_ID")
	secretID  = os.Getenv("SECRET_ID")
)

func main() {
	// app := cli.NewApp()
	// app.Name = "Espresso Keystore and Secrets Manager"
	// app.Usage = "Update a Secret Manager secret with Sequencer private keys and DB keys."
	// app.Version = version
	// app.Copyright = "(c) 2024 Nethermind"

	ctx := context.Background()
	cmd := &cli.Command{
		Name:    "espresso-cli",
		Version: version,
		Usage:   "Update a Secret Manager secret with Sequencer private keys and DB keys.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "project-id",
				Usage:       "Google Cloud Project ID",
				Destination: &projectID,
			},
			&cli.StringFlag{
				Name:        "secret-id",
				Usage:       "Secret Manager secret ID",
				Destination: &secretID,
			},
		},
		Commands: []*cli.Command{
			keystoreCMD(ctx),
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}

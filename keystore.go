package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/urfave/cli/v2"
)

var (
	keysPath string
)

func keystoreCMD() *cli.Command {
	cmd := &cli.Command{
		Name:  "pv-keys",
		Usage: "Update a Secret Manager secret with Sequencer private keys.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "keys-path",
				Usage:       "Path to the directory containing the .env files with the private keys",
				Destination: &keysPath,
				DefaultText: "/keys",
				EnvVars:     []string{"KEYS_PATH"},
			},
		},
		Action: keystoreAction,
	}

	return cmd
}

func keystoreAction(cCtx *cli.Context) error {
	ctx := cCtx.Context

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to setup client: %v", err)
	}
	defer client.Close()

	// Iterate over all .env files in the directory
	files, err := os.ReadDir(keysPath)
	if err != nil {
		return fmt.Errorf("Failed to read directory: %v", err)
	}

	// Existing secrets will be stored in this map
	secrets := make(map[string]string)

	// Load current secret contents
	secretData, err := getSecret(ctx, client, projectID, secretID, "latest")
	if err != nil {
		return fmt.Errorf("Failed to access initial secret: %v", err)
	}

	// Parse the existing secret to avoid duplications
	lines := strings.Split(string(secretData), "\n")
	for _, line := range lines {
		if keyValue := strings.SplitN(line, "=", 2); len(keyValue) == 2 {
			secrets[keyValue[0]] = keyValue[1]
		}
	}

	// Process each .env file
	updated := false
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".env" {
			contents, err := os.ReadFile(filepath.Join(keysPath, file.Name()))
			if err != nil {
				log.Printf("Error reading file %s: %v", file.Name(), err)
				continue
			}

			// Index is assumed from file name, extracting from 'x.env'
			index := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			lines := strings.Split(string(contents), "\n")
			for _, line := range lines {
				if keyValue := strings.SplitN(line, "=", 2); len(keyValue) == 2 {
					key := fmt.Sprintf("%s_%s", keyValue[0], index)
					if _, exists := secrets[key]; !exists {
						secrets[key] = keyValue[1]
						updated = true
					}
				}
			}
		}
	}
	if updated {
		// Process and update the secret
		err = updateSecret(ctx, client, projectID, secretID, secrets)
		if err != nil {
			log.Printf("Error updating secret: %v", err)
		}
		log.Printf("Secrets updated")
	} else {
		log.Printf("No new secrets to update")
		return nil
	}

	// Sleep for 5 seconds to allow the secret to propagate
	log.Printf("Sleeping for 5 seconds to allow the secret to propagate")
	<-time.After(5 * time.Second)

	// Reload and print final secret contents
	// finalSecretData, err := getSecret(ctx, client, projectID, secretID, "latest")
	// if err != nil {
	// 	return fmt.Errorf("Failed to access final secret: %v", err)
	// }
	// log.Printf("Final Secret Contents:\n%s", string(finalSecretData))
	return nil
}

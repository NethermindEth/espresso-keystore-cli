package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/urfave/cli/v3"
)

var (
	dbHost = os.Getenv("SEQUENCER_POSTGRES_HOST")
	dbUser = os.Getenv("SEQUENCER_POSTGRES_USER")
	dbPass = os.Getenv("SEQUENCER_POSTGRES_PASS")

	dbHostKey = "ESPRESSO_SEQUENCER_POSTGRES_HOST"
	dbUserKey = "ESPRESSO_SEQUENCER_POSTGRES_USER"
	dbPassKey = "ESPRESSO_SEQUENCER_POSTGRES_PASS"
)

type dbKey struct {
	Key   string
	Value string
}

func dbKeysCMD(ctx context.Context) *cli.Command {
	cmd := &cli.Command{
		Name:  "db-keys",
		Usage: "Update a Secret Manager secret with DB keys.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "db-host",
				Usage:       "Database host URL",
				Destination: &dbHost,
			},
			&cli.StringFlag{
				Name:        "db-user",
				Usage:       "Database username",
				Destination: &dbUser,
			},
			&cli.StringFlag{
				Name:        "db-pass",
				Usage:       "Database password",
				Destination: &dbPass,
			},
		},
		Action: dbKeysAction,
	}

	return cmd
}

func dbKeysAction(ctx context.Context, cmd *cli.Command) error {
	// Validate required flags
	if err := validateRequiredOptions(); err != nil {
		return err
	}
	if dbHost == "" {
		return fmt.Errorf("Database host URL is required")
	}
	if dbUser == "" {
		return fmt.Errorf("Database username is required")
	}
	if dbPass == "" {
		return fmt.Errorf("Database password is required")
	}

	// Initialize keys
	keys := []dbKey{
		{Key: dbHostKey, Value: dbHost},
		{Key: dbUserKey, Value: dbUser},
		{Key: dbPassKey, Value: dbPass},
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to setup client: %v", err)
	}
	defer client.Close()

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

	updated := false
	for _, key := range keys {
		_, exists := secrets[key.Key]
		if !exists {
			secrets[key.Key] = key.Value
			updated = true
		} else if secrets[key.Key] != key.Value {
			secrets[key.Key] = key.Value
			updated = true
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
	finalSecretData, err := getSecret(ctx, client, projectID, secretID, "latest")
	if err != nil {
		return fmt.Errorf("Failed to access final secret: %v", err)
	}
	log.Printf("Final Secret Contents:\n%s", string(finalSecretData))
	return nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func main() {
	// Set up the context and client for Secret Manager
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to setup client: %v", err)
	}
	defer client.Close()

	// Get environment variables from Pod spec.
	keysPath := os.Getenv("KEYS_PATH")
	projectID := os.Getenv("PROJECT_ID")
	secretID := os.Getenv("SECRET_ID")

	// Iterate over all .env files in the directory
	files, err := os.ReadDir(keysPath)
	if err != nil {
		log.Fatalf("Failed to read directory: %v", err)
	}

	// Existing secrets will be stored in this map
	secrets := make(map[string]string)

	// Load current secret contents
	secretData, err := getSecret(ctx, client, projectID, secretID, "latest")
	if err != nil {
		log.Fatalf("Failed to access initial secret: %v", err)
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
	}

	// Sleep for 5 seconds to allow the secret to propagate
	log.Printf("Sleeping for 5 seconds to allow the secret to propagate")
	<-time.After(5 * time.Second)

	// Reload and print final secret contents
	finalSecretData, err := getSecret(ctx, client, projectID, secretID, "latest")
	if err != nil {
		log.Fatalf("Failed to access final secret: %v", err)
	}
	log.Printf("Final Secret Contents:\n%s", string(finalSecretData))
}

func getSecret(ctx context.Context, client *secretmanager.Client, projectID, secretID, version string) ([]byte, error) {
	accessReq := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/%s", projectID, secretID, version),
	}
	result, err := client.AccessSecretVersion(ctx, accessReq)
	if err != nil {
		return nil, err
	}
	return result.Payload.Data, nil
}

func updateSecret(ctx context.Context, client *secretmanager.Client, projectID, secretID string, secrets map[string]string) error {
	var newSecretValue string

	for key, value := range secrets {
		newSecretValue += fmt.Sprintf("%s=%s\n", key, value)
	}

	secretVersionRequest := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", projectID, secretID),
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(newSecretValue),
		},
	}
	_, err := client.AddSecretVersion(ctx, secretVersionRequest)
	if err != nil {
		return err
	}
	return nil
}

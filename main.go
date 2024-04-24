package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	existingSecrets := make(map[string]string)

	// Load current secret contents
	secretData, err := getSecret(ctx, client, projectID, secretID, "latest")
	if err != nil {
		log.Fatalf("Failed to access initial secret: %v", err)
	}

	// Parse the existing secret to avoid duplications
	lines := strings.Split(string(secretData), "\n")
	for _, line := range lines {
		if keyValue := strings.SplitN(line, "=", 2); len(keyValue) == 2 {
			existingSecrets[keyValue[0]] = keyValue[1]
		}
	}

	// Process each .env file
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".env" {
			contents, err := os.ReadFile(filepath.Join(keysPath, file.Name()))
			if err != nil {
				log.Printf("Error reading file %s: %v", file.Name(), err)
				continue
			}

			// Index is assumed from file name, extracting from 'x.env'
			index := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

			// Process and update the secret
			err = processAndUpdateSecret(ctx, client, projectID, secretID, string(contents), index, existingSecrets)
			if err != nil {
				log.Printf("Error processing file %s: %v", file.Name(), err)
			}
		}
	}

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

func processAndUpdateSecret(ctx context.Context, client *secretmanager.Client, projectID, secretID, contents, index string, existingSecrets map[string]string) error {
	lines := strings.Split(contents, "\n")
	updated := false
	var newSecretValue string

	for _, line := range lines {
		if keyValue := strings.SplitN(line, "=", 2); len(keyValue) == 2 {
			key := fmt.Sprintf("%s_%s", keyValue[0], index)
			if _, exists := existingSecrets[key]; !exists {
				newSecretValue += fmt.Sprintf("%s=%s\n", key, keyValue[1])
				updated = true
			}
		}
	}

	if updated {
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
	}
	return nil
}

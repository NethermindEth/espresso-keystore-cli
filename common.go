package main

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

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

func validateRequiredOptions() error {
	var errMsg string
	if projectID == "" {
		errMsg += "Project ID is required. "
	}
	if secretID == "" {
		errMsg += "Secret ID is required. "
	}

	return fmt.Errorf(errMsg)
}

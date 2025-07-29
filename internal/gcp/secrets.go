package gcp

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretManager handles Google Cloud Secret Manager operations
type SecretManager struct {
	client    *secretmanager.Client
	projectID string
}

// NewSecretManager creates a new Secret Manager client
func NewSecretManager(ctx context.Context, projectID string) (*SecretManager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &SecretManager{
		client:    client,
		projectID: projectID,
	}, nil
}

// GetSecret retrieves a secret value from Secret Manager
func (sm *SecretManager) GetSecret(ctx context.Context, secretName string) (string, error) {
	// Build the resource name
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sm.projectID, secretName)

	// Access the secret version
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
}

// Close closes the Secret Manager client
func (sm *SecretManager) Close() error {
	return sm.client.Close()
}

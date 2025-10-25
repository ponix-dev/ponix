package ttn

import (
	"github.com/ponix-dev/ponix/internal/conf"
)

const (
	// ApiKeyEnvVar is the environment variable name for the TTN API key.
	ApiKeyEnvVar = "TTN_API_KEY"
	// ApiCollaboratorEnvVar is the environment variable name for the TTN API collaborator identifier.
	ApiCollaboratorEnvVar = "TTN_API_COLLABORATOR"
	// TTNApplicationEnvVar is the environment variable name for the TTN application ID.
	TTNApplicationEnvVar = "TTN_APPLICATION"
)

// TTNApplicationFromEnv retrieves the TTN application ID from the TTN_APPLICATION environment variable.
func TTNApplicationFromEnv() (string, error) {
	return conf.FromEnv(TTNApplicationEnvVar)
}

// ApiKeyFromEnv retrieves the TTN API key from the TTN_API_KEY environment variable.
func ApiKeyFromEnv() (string, error) {
	return conf.FromEnv(ApiKeyEnvVar)
}

// ApiCollaboratorFromEnv retrieves the TTN API collaborator identifier from the TTN_API_COLLABORATOR environment variable.
func ApiCollaboratorFromEnv() (string, error) {
	return conf.FromEnv(ApiCollaboratorEnvVar)
}

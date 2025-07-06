package ttn

import (
	"github.com/ponix-dev/ponix/internal/conf"
)

const (
	ApiKeyEnvVar          = "TTN_API_KEY"
	ApiCollaboratorEnvVar = "TTN_API_COLLABORATOR"
)

func ApiKeyFromEnv() (string, error) {
	return conf.FromEnv(ApiKeyEnvVar)
}

func ApiCollaboratorFromEnv() (string, error) {
	return conf.FromEnv(ApiCollaboratorEnvVar)
}

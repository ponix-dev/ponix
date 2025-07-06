package ory

import (
	ory "github.com/ory/client-go"
)

type OryConfigOption func(*ory.Configuration)

func WithOryServerConfiguration(sc ory.ServerConfiguration) OryConfigOption {
	return func(c *ory.Configuration) {
		c.Servers = append(c.Servers, sc)
	}
}

func NewApiClient(options ...OryConfigOption) *ory.APIClient {
	// TODO: figure out a better default
	configuration := ory.NewConfiguration()
	configuration.Servers = []ory.ServerConfiguration{}

	for _, option := range options {
		option(configuration)
	}

	ac := ory.NewAPIClient(configuration)

	return ac
}

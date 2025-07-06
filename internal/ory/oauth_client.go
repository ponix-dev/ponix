package ory

import (
	"context"
	"encoding/json"
	"fmt"

	ory "github.com/ory/client-go"
	"github.com/ponix-dev/ponix/internal/auth"
	"github.com/ponix-dev/ponix/internal/telemetry/stacktrace"
)

type OauthClient struct {
	apiClient *ory.APIClient
}

func NewOauthClient(ac *ory.APIClient) *OauthClient {
	return &OauthClient{
		apiClient: ac,
	}
}

func (oac *OauthClient) CreateOAuth2Client(ctx context.Context, client auth.OauthClient) (auth.OauthClient, error) {
	oAuth2Client := *ory.NewOAuth2Client()
	oAuth2Client.SetClientName(client.Name)

	_, _, err := oac.apiClient.OAuth2API.CreateOAuth2Client(ctx).OAuth2Client(oAuth2Client).Execute()
	if err != nil {
		return auth.OauthClient{}, stacktrace.NewStackTraceError(err)
	}

	return client, nil
}

func (oac *OauthClient) ListOAuth2Clients(ctx context.Context) ([]auth.OauthClient, error) {
	clients, _, err := oac.apiClient.OAuth2API.ListOAuth2Clients(ctx).Execute()
	if err != nil {
		return nil, stacktrace.NewStackTraceError(err)
	}

	ocs := make([]auth.OauthClient, len(clients))

	for i, c := range clients {
		oc := auth.OauthClient{
			Name: c.GetClientName(),
		}

		ocs[i] = oc

		out, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return nil, stacktrace.NewStackTraceError(err)
		}

		fmt.Println(string(out))
	}

	return ocs, nil
}

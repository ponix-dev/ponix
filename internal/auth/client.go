package auth

import "context"

type OauthClientStorer interface {
	CreateOAuth2Client(ctx context.Context, client OauthClient) (OauthClient, error)
	ListOAuth2Clients(ctx context.Context) ([]OauthClient, error)
}

type AuthClient struct {
	oauthClientStore OauthClientStorer
}

func NewAuthClient(ocs OauthClientStorer) *AuthClient {
	return &AuthClient{
		oauthClientStore: ocs,
	}
}

func (ac *AuthClient) CreateOauth2Client(ctx context.Context, client OauthClient) (OauthClient, error) {
	return ac.oauthClientStore.CreateOAuth2Client(ctx, client)
}

func (ac *AuthClient) ListOauth2Clients(ctx context.Context) ([]OauthClient, error) {
	return ac.oauthClientStore.ListOAuth2Clients(ctx)
}

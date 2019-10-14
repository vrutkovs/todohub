package github

import (
	"context"
	api "github.com/google/go-github/v28/github"
	"github.com/vrutkovs/trellohub/pkg/trello"
	"golang.org/x/oauth2"
)

type Client struct {
	api      *api.Client
	settings GithubSettings
}

func GetClient(settings GithubSettings) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: settings.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		api:      api.NewClient(tc),
		settings: settings,
	}
}

func (c *Client) UpdateTrello(tr *trello.Client) {
	opt := &api.RepositoryListByOrgOptions{Type: "public"}
	c.api.Repositories.ListByOrg(context.Background(), "github", opt)
}

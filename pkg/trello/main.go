package trello

import (
	api "github.com/adlio/trello"
)

type Client struct {
	api      *api.Client
	settings TrelloSettings
}

func GetClient(settings TrelloSettings) *Client {
	return &Client{
		api:      api.NewClient(settings.AppKey, settings.Token),
		settings: settings,
	}
}

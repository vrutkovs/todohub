package trello

import "github.com/adlio/trello"

func getClient() *trello.Client {
	return trello.NewClient("foo", "bar")
}

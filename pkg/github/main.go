package github

import (
	"context"
	"fmt"
	retry "github.com/avast/retry-go"
	api "github.com/google/go-github/v28/github"
	"github.com/vrutkovs/trellohub/pkg/trello"
	"golang.org/x/oauth2"
	"log"
)

type Client struct {
	api      *api.Client
	settings GithubSettings
}

type IssueInfo struct {
	title string
	url   string
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
	log.Println("github: updating trello")
	for list, searchQuery := range c.settings.GithubSearchList {
		query := searchQuery
		if len(c.settings.SearchPrefix) > 0 {
			query = fmt.Sprintf("%s %s", c.settings.SearchPrefix, searchQuery)
		}
		searchResults, err := c.getIssueInfoForSearchQuery(query)
		if err != nil {
			panic(err)
		}
		log.Printf("github: fetched search results for list '%s'", list)

		listID, err := tr.EnsureListExists(list)
		if err != nil {
			panic(err)
		}
		log.Printf("github: got list ID %s", listID)

		for _, item := range searchResults {
			card, err := tr.AddItemToList(item.title, listID)
			if err != nil {
				panic(err)
			}
			log.Printf("github: created item %s", item.title)
			err = tr.AttachLink(card, item.url)
			if err != nil {
				panic(err)
			}
			log.Printf("github: attached url %s", item.url)
		}
	}

	opt := &api.RepositoryListByOrgOptions{Type: "public"}
	c.api.Repositories.ListByOrg(context.Background(), "github", opt)
}

func (c *Client) getIssueInfoForSearchQuery(searchQuery string) ([]IssueInfo, error) {
	ctx := context.Background()
	opts := &api.SearchOptions{Sort: "created", Order: "asc"}
	results := make([]IssueInfo, 0)
	err := retry.Do(
		func() error {
			result, _, err := c.api.Search.Issues(ctx, searchQuery, opts)
			if err != nil {
				return err
			}
			for _, issue := range result.Issues {
				ii := IssueInfo{
					title: issue.GetTitle(),
					url:   issue.GetHTMLURL(),
				}
				results = append(results, ii)
			}
			return nil
		},
		retry.RetryIf(func(err error) bool {
			return !isCritical(err)
		}),
	)
	return results, err
}

// Check if github error is fatal
func isCritical(err error) bool {
	if _, ok := err.(*api.RateLimitError); ok {
		log.Println("hit rate limit")
		return false
	}
	return true
}

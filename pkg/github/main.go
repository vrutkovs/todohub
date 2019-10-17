package github

import (
	"context"
	"fmt"
	retry "github.com/avast/retry-go"
	api "github.com/google/go-github/v28/github"
	"github.com/vrutkovs/trellohub/pkg/trello"
	"golang.org/x/oauth2"
	"log"
	"sync"
)

type Client struct {
	api      *api.Client
	settings GithubSettings
}

type IssueInfo struct {
	title string
	url   string
}

type WorkerData struct {
	query string
	list  string
	tr    *trello.Client
}

const PARALLEL_WORKERS = 5

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

func (c *Client) githubWorker(wData WorkerData, wg *sync.WaitGroup) {
	defer wg.Done()

	query := wData.query
	if len(c.settings.SearchPrefix) > 0 {
		query = fmt.Sprintf("%s %s", c.settings.SearchPrefix, wData.query)
	}
	searchResults, err := c.getIssueInfoForSearchQuery(query)
	if err != nil {
		panic(err)
	}
	log.Printf("github: fetched search results for list '%s'", wData.list)

	listID, err := wData.tr.EnsureListExists(wData.list)
	if err != nil {
		panic(err)
	}
	log.Printf("github: got list ID %s", listID)

	log.Println("github: fetching existing cards")
	cardsToRemove, err := wData.tr.FetchCardsInList(listID)
	if err != nil {
		panic(err)
	}

	log.Println("github: adding new cards")
	for _, item := range searchResults {

		if _, ok := cardsToRemove[item.title]; ok {
			log.Printf("github: found existing card %s", item.title)
			continue
		}

		card, err := wData.tr.AddItemToList(item.title, listID)
		if err != nil {
			panic(err)
		}
		log.Printf("github: created item %s", item.title)
		err = wData.tr.AttachLink(card, item.url)
		if err != nil {
			panic(err)
		}

		delete(cardsToRemove, item.title)
	}

	log.Println("github: looking for old cards")
	for _, id := range cardsToRemove {
		wData.tr.CloseCard(id)
	}
}

func (c *Client) UpdateTrello(tr *trello.Client) {
	if c.settings.BoardID != "" {
		tr.SetBoardID(c.settings.BoardID)
	}

	var wg sync.WaitGroup

	log.Println("github: updating trello")
	for list, searchQuery := range c.settings.GithubSearchList {
		workerData := WorkerData{
			query: searchQuery,
			list:  list,
			tr:    tr,
		}
		wg.Add(1)
		go c.githubWorker(workerData, &wg)
	}
	wg.Wait()

	log.Println("github update completed")
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

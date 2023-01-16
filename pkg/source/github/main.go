package github

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/avast/retry-go"
	api "github.com/google/go-github/v28/github"
	"github.com/vrutkovs/todohub/pkg/issue"
	"github.com/vrutkovs/todohub/pkg/storage"
	"golang.org/x/oauth2"
)

// ParallelWorkers is a number of worker threads run in parallel
const ParallelWorkers = 1

// Client holds information about github client
type Client struct {
	api       *api.Client
	storage   *storage.Client
	settings  *Settings
	issueList GithubIssueList
}

// New returns github client
func New(s *Settings, storage *storage.Client) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		api:      api.NewClient(tc),
		storage:  storage,
		settings: s,
	}
}

// GithubIssue implements source.Issue
type GithubIssue struct {
	title string
	url   string
}

func (i GithubIssue) Title() string {
	return i.title
}

func (i GithubIssue) Url() string {
	return i.url
}

// GithubIssueList implements source.IssueList
type GithubIssueList struct {
	issues map[string][]GithubIssue
}

func (i GithubIssueList) Name(name string) []GithubIssue {
	if list, ok := i.issues[name]; ok {
		return list
	}
	return nil
}

func (s Client) Issues() GithubIssueList {
	return s.issueList
}

// WorkerData holds info about worker payload
type WorkerData struct {
	project string
	query   string
	storage storage.Client
}

// githubWorker runs queries in github
func (c *Client) githubWorker(wData WorkerData, wg *sync.WaitGroup) {
	defer wg.Done()

	// Run the query
	query := wData.query
	if len(c.settings.SearchPrefix) > 0 {
		query = fmt.Sprintf("%s %s", c.settings.SearchPrefix, wData.query)
	}
	searchResults, err := c.getIssueInfoForSearchQuery(query)
	if err != nil {
		panic(err)
	}
	// log.Printf("github: fetched search results for project '%s'", wData.project)
	// Build a new list of issues from search results
	required := issue.IssueList{
		Issues: make([]issue.Issue, len(searchResults)),
	}
	for i, issue := range searchResults {
		required.Issues[i] = GithubIssue{
			title: issue.title,
			url:   issue.url,
		}
	}

	// Create a list if its missing
	// log.Println("github: fetching existing cards")
	err = wData.storage.CreateProject(wData.project)
	if err != nil {
		panic(err)
	}

	// Fetch existing cards and mark all cards for removal
	existingIssues, err := wData.storage.GetIssues(wData.project)
	if err != nil {
		panic(err)
	}
	existing := issue.IssueList{
		Issues: make([]issue.Issue, len(existingIssues)),
	}

	// Drop internal values to make intersection work
	for i, issue := range existingIssues {
		existing.Issues[i] = GithubIssue{
			title: issue.Title(),
			url:   issue.Url(),
		}
	}

	titleOnlyComparison := (*c.storage).CompareByTitleOnly()

	// Create an intersection from these two lists
	hashExisting := existing.MakeHashList(titleOnlyComparison)
	hashRequired := required.MakeHashList(titleOnlyComparison)
	// Remove all cards in existing which are not in intersection
	// log.Println("github: removing old cards")
	for _, el := range issue.OuterSection(hashExisting, hashRequired).Issues {
		err := wData.storage.Delete(wData.project, el)
		if err != nil {
			panic(err)
		}
		log.Printf("github: removed item %s", el.Title())
	}

	// Add all cards from required which are not in intersection
	// log.Println("github: adding new cards")
	for _, i := range issue.OuterSection(hashRequired, hashExisting).Issues {
		err := wData.storage.Create(wData.project, i)
		if err != nil {
			panic(err)
		}
		log.Printf("github: created item %s", i.Title())
	}
	wData.storage.Sync()
}

// UpdateTrello runs search queries and applies changes in trello
func (c *Client) Sync() error {
	var wg sync.WaitGroup
	storageClient := *c.storage

	// log.Println("github: updating storage")
	for project, query := range c.settings.SearchList {
		workerData := WorkerData{
			project: project,
			query:   query,
			storage: storageClient,
		}
		wg.Add(1)
		//go c.githubWorker(workerData, &wg)
		c.githubWorker(workerData, &wg)
	}
	wg.Wait()
	// log.Println("github update completed")
	return storageClient.Sync()

}

// getIssueInfoForSearchQuery runs the query and returns a list of issues
func (c *Client) getIssueInfoForSearchQuery(searchQuery string) ([]GithubIssue, error) {
	ctx := context.Background()
	opts := &api.SearchOptions{Sort: "created", Order: "asc"}
	results := make([]GithubIssue, 0)
	err := retry.Do(
		func() error {
			result, _, err := c.api.Search.Issues(ctx, searchQuery, opts)
			if err != nil {
				return err
			}
			for _, issue := range result.Issues {
				ii := GithubIssue{
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

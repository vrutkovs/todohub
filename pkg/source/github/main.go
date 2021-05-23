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
const ParallelWorkers = 5

// Client holds information about github client
type Client struct {
	api       *api.Client
	storage   *storage.Client
	settings  Settings
	issueList GithubIssueList
}

// New returns github client
func (c *Client) New(s Settings, storage *storage.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.settings.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	c.api = api.NewClient(tc)
	c.storage = storage
	c.settings = s
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
	if len(c.settings.searchPrefix) > 0 {
		query = fmt.Sprintf("%s %s", c.settings.searchPrefix, wData.query)
	}
	searchResults, err := c.getIssueInfoForSearchQuery(query)
	if err != nil {
		panic(err)
	}
	log.Printf("github: fetched search results for project '%s'", wData.project)

	// Create a list if its missing
	log.Println("github: fetching existing cards")
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
		Issues: existingIssues,
	}

	// Build a new list of issues from search results
	required := issue.IssueList{
		Issues: make([]issue.Issue, len(searchResults)-1),
	}
	for _, issue := range searchResults {
		required.Issues = append(required.Issues, GithubIssue{
			title: issue.title,
			url:   issue.url,
		})
	}

	// Create an intersection from these two lists
	interfaceIntersection := required.InterSection(existing)
	intersection := issue.IssueList{
		Issues: make([]issue.Issue, len(interfaceIntersection)-1),
	}
	for _, item := range interfaceIntersection {
		i := item.(issue.Issue)
		intersection.Issues = append(intersection.Issues, i)
	}

	// Add all cards from required which are not in intersection
	log.Println("github: adding new cards")
	for _, i := range required.Issues {
		if _, ok := intersection.Get(i.Title()); !ok {
			err := wData.storage.Create(wData.project, i)
			if err != nil {
				panic(err)
			}
			log.Printf("github: created item %s", i.Title())
		}
	}

	// Remove all cards in existing which are not in intersection
	log.Println("github: removing old cards")
	for _, i := range existing.Issues {
		if _, ok := intersection.Get(i.Title()); !ok {
			err := wData.storage.Delete(wData.project, i)
			if err != nil {
				panic(err)
			}
			log.Printf("github: removed item %s", i.Title())
		}
	}
}

// UpdateTrello runs search queries and applies changes in trello
func (c *Client) Sync() {
	var wg sync.WaitGroup

	log.Println("github: updating trello")
	for project, query := range c.settings.searchList {
		workerData := WorkerData{
			project: project,
			query:   query,
			storage: *c.storage,
		}
		wg.Add(1)
		go c.githubWorker(workerData, &wg)
	}
	wg.Wait()

	log.Println("github update completed")
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

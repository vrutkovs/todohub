package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/avast/retry-go"
	api "github.com/google/go-github/v28/github"
	"github.com/sirupsen/logrus"
	"github.com/vrutkovs/todohub/pkg/issue"
	"github.com/vrutkovs/todohub/pkg/storage"
	"golang.org/x/oauth2"
)

// ParallelWorkers is a number of worker threads run in parallel.
const ParallelWorkers = 1

// Client holds information about github client.
type Client struct {
	api           *api.Client
	storageClient *storage.Client
	settings      *Settings
	issueList     IssueList
	logger        *logrus.Logger
}

// New returns github client.
func New(s *Settings, storageClient storage.Client, logger *logrus.Logger) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		api:           api.NewClient(tc),
		storageClient: &storageClient,
		settings:      s,
		logger:        logger,
	}
}

// Issue implements source.Issue.
type Issue struct {
	title string
	url   string
	repo  string
}

func (i Issue) Title() string {
	return i.title
}

func (i Issue) URL() string {
	return i.url
}

func (i Issue) Repo() string {
	return i.repo
}

// IssueList implements source.IssueList.
type IssueList struct {
	issues map[string][]Issue
}

func (i IssueList) Name(name string) []Issue {
	if list, ok := i.issues[name]; ok {
		return list
	}
	return nil
}

func (c Client) Issues() IssueList {
	return c.issueList
}

// WorkerData holds info about worker payload.
type WorkerData struct {
	project string
	query   string
	storage storage.Client
}

// githubWorker runs queries in github.
func (c *Client) githubWorker(wData WorkerData, wg *sync.WaitGroup) {
	logger := c.logger.WithFields(logrus.Fields{"source": "github", "project": wData.project})
	defer wg.Done()

	// Run the query
	query := wData.query
	if c.settings.SearchPrefix != "" {
		query = fmt.Sprintf("%s %s", c.settings.SearchPrefix, wData.query)
	}
	searchResults, err := c.getIssueInfoForSearchQuery(query)
	if err != nil {
		c.logger.Fatal(err)
	}
	logger.Info("fetched search results")
	// Build a new list of issues from search results
	required := issue.List{
		Issues: make([]issue.Issue, len(searchResults)),
	}
	for i, issue := range searchResults {
		required.Issues[i] = Issue{
			title: issue.title,
			url:   issue.url,
			repo:  issue.repo,
		}
	}

	// Create a list if its missing
	logger.Info("fetching existing cards")
	err = wData.storage.CreateProject(wData.project)
	if err != nil {
		c.logger.Fatal(err)
	}

	// Fetch existing cards and mark all cards for removal
	existingIssues, err := wData.storage.GetIssues(wData.project)
	if err != nil {
		c.logger.Fatal(err)
	}
	existing := issue.List{
		Issues: make([]issue.Issue, len(existingIssues)),
	}

	// Drop internal values to make intersection work
	for i, issue := range existingIssues {
		existing.Issues[i] = Issue{
			title: issue.Title(),
			url:   issue.URL(),
			repo:  issue.Repo(),
		}
	}

	titleOnlyComparison := (*c.storageClient).CompareByTitleOnly()

	// Create an intersection from these two lists
	hashExisting := existing.MakeHashList(titleOnlyComparison)
	hashRequired := required.MakeHashList(titleOnlyComparison)
	// Remove all cards in existing which are not in intersection
	logger.Info("removing old cards")
	for _, el := range issue.OuterSection(hashExisting, hashRequired).Issues {
		err := wData.storage.Delete(wData.project, el)
		if err != nil {
			c.logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"item": el.Title()}).Info("removing old card")
	}

	// Add all cards from required which are not in intersection
	logger.Info("adding new cards")
	for _, i := range issue.OuterSection(hashRequired, hashExisting).Issues {
		err := wData.storage.Create(wData.project, i)
		if err != nil {
			c.logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"item": i.Title()}).Info("created card")
	}
	if err := wData.storage.Sync(query); err != nil {
		c.logger.Fatal(err)
	}
	logger.Info("done")
}

// Sync runs search queries and applies changes in storage.
func (c *Client) Sync(description string) error {
	var wg sync.WaitGroup
	storageClient := *c.storageClient

	logger := c.logger.WithFields(logrus.Fields{"source": "github", "description": description})
	logger.Info("syncing")
	for project, query := range c.settings.SearchList {
		logger.WithField("project", project).Info("started")
		workerData := WorkerData{
			project: project,
			query:   query,
			storage: storageClient,
		}
		wg.Add(1)

		c.githubWorker(workerData, &wg)
		logger.WithField("project", project).Info("done")

	}
	wg.Wait()
	logger.Info("sync completed")
	return nil
}

// getIssueInfoForSearchQuery runs the query and returns a list of issues.
func (c *Client) getIssueInfoForSearchQuery(searchQuery string) ([]Issue, error) {
	ctx := context.Background()
	opts := &api.SearchOptions{Sort: "created", Order: "asc"}
	results := make([]Issue, 0)
	err := retry.Do(
		func() error {
			result, _, err := c.api.Search.Issues(ctx, searchQuery, opts)
			if err != nil {
				c.logger.WithFields(logrus.Fields{"source": "github", "query": searchQuery}).WithError(err)
				return err
			}
			for _, issue := range result.Issues {
				ii := Issue{
					title: issue.GetTitle(),
					url:   issue.GetHTMLURL(),
					repo:  repoSlug(issue.GetRepositoryURL()),
				}
				results = append(results, ii)
			}
			return nil
		},
		retry.RetryIf(func(err error) bool {
			var errRateLimit *api.RateLimitError
			return errors.As(err, &errRateLimit)
		}),
	)
	c.logger.
		WithFields(logrus.Fields{"source": "github", "query": searchQuery, "count": len(results)}).
		WithError(err).
		Info("results fetched")
	return results, err
}

// Build repo slug from Repository.
func repoSlug(repoURL string) string {
	splitString := strings.Split(repoURL, "/")
	if len(splitString) < 4 {
		return ""
	}
	return fmt.Sprintf("%s/%s", splitString[len(splitString)-2], splitString[len(splitString)-1])
}

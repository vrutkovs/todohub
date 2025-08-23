package jira

import (
	"context"
	"sync"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"github.com/vrutkovs/todohub/pkg/issue"
	"github.com/vrutkovs/todohub/pkg/storage"
)

// Client holds information about jira client.
type Client struct {
	api           *jira.Client
	storageClient *storage.Client
	settings      *Settings
	issueList     IssueList
	logger        *logrus.Logger
}

// WorkerData holds info about worker payload.
type WorkerData struct {
	project string
	query   string
	storage storage.Client
}

// Issue implements source.Issue.
type Issue struct {
	title   string
	url     string
	project string
}

func (i Issue) Title() string {
	return i.title
}

func (i Issue) URL() string {
	return i.url
}

func (i Issue) Repo() string {
	return i.project
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

// New returns jira client.
func New(s *Settings, storageClient storage.Client, logger *logrus.Logger) (*Client, error) {
	tp := jira.BearerAuthTransport{
		Token: s.Token,
	}
	client, err := jira.NewClient(s.Endpoint, tp.Client())
	if err != nil {
		return nil, err
	}
	return &Client{
		api:           client,
		storageClient: &storageClient,
		settings:      s,
		logger:        logger,
	}, nil
}

// Sync runs search queries and applies changes in storage.
func (c *Client) Sync(description string) error {
	var wg sync.WaitGroup
	storageClient := *c.storageClient

	logger := c.logger.WithFields(logrus.Fields{"source": "jira", "desc": description})

	logger.Info("syncing")
	for project, query := range c.settings.SearchList {
		workerData := WorkerData{
			project: project,
			query:   query,
			storage: storageClient,
		}
		wg.Add(1)

		c.jiraWorker(workerData, &wg)
	}
	wg.Wait()
	logger.Info("syncing done")
	return nil
}

// jiraWorker runs queries in jira.
func (c *Client) jiraWorker(wData WorkerData, wg *sync.WaitGroup) {
	logger := c.logger.WithFields(logrus.Fields{"source": "jira", "project": wData.project})
	defer wg.Done()

	// Run the query
	query := wData.query
	searchResults, err := c.getIssueInfoForSearchQuery(query)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("fetched search results")
	// log.Printf("github: fetched search results for project '%s'", wData.project)
	// Build a new list of issues from search results
	required := issue.List{
		Issues: make([]issue.Issue, len(searchResults)),
	}
	for i, issue := range searchResults {
		required.Issues[i] = Issue{
			title:   issue.title,
			url:     issue.url,
			project: issue.project,
		}
	}

	// Create a list if its missing
	logger.Info("fetching existing cards")
	err = wData.storage.CreateProject(wData.project)
	if err != nil {
		logger.Fatal(err)
	}

	// Fetch existing cards and mark all cards for removal
	existingIssues, err := wData.storage.GetIssues(wData.project)
	if err != nil {
		logger.Fatal(err)
	}
	logger.WithField("count", len(existingIssues)).Info("fetched existing cards")
	existing := issue.List{
		Issues: make([]issue.Issue, len(existingIssues)),
	}

	// Drop internal values to make intersection work
	for i, issue := range existingIssues {
		existing.Issues[i] = Issue{
			title:   issue.Title(),
			url:     issue.URL(),
			project: issue.Repo(),
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
			logger.Fatal(err)
		}
		logger.WithField("item", el.Title()).Info("removed")
	}

	// Add all cards from required which are not in intersection
	logger.Info("adding new cards")
	for _, i := range issue.OuterSection(hashRequired, hashExisting).Issues {
		err := wData.storage.Create(wData.project, i)
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithField("item", i.Title()).Info("created")
	}
	if err := wData.storage.Sync(query); err != nil {
		logger.Fatal(err)
	}
}

// getIssueInfoForSearchQuery runs the query and returns a list of issues.
func (c *Client) getIssueInfoForSearchQuery(searchQuery string) ([]Issue, error) {
	logger := c.logger.WithFields(logrus.Fields{"source": "jira", "query": searchQuery})
	logger.Info("starting")

	ctx := context.Background()
	results := make([]Issue, 0)
	err := retry.Do(
		func() error {
			appendFunc := func(i jira.Issue) (err error) {
				result := Issue{
					title:   i.Fields.Summary,
					url:     c.buildJiraTicketUrl(i.Key),
					project: i.Fields.Project.Key,
				}
				results = append(results, result)
				return nil
			}
			err := c.api.Issue.SearchPages(ctx, searchQuery, nil, appendFunc)
			logger.WithError(err).Info("done")
			return err
		},
	)
	logger.WithError(err).WithField("count", len(results)).Info("results fetched")
	return results, err
}

func (c *Client) buildJiraTicketUrl(key string) string {
	return c.api.BaseURL.JoinPath("browse", key).String()
}

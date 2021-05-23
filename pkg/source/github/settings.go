package github

import (
	"context"

	api "github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

// Settings stores info about github connection
type Settings struct {
	token        string          `yaml:"token"`
	boardID      string          `yaml:"boardid,omitempty"`
	searchPrefix string          `yaml:"search_prefix,omitempty"`
	searchList   GithubIssueList `yaml:"lists"`
}

// Implement source.Settings
func (s Settings) ID() string {
	return "github"
}

func (s Settings) Self() interface{} {
	return &s
}

// Authenticate returns github client
func (s Settings) Authenticate() interface{} {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: s.token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		api:      api.NewClient(tc),
		settings: s,
	}
}

func (s Settings) Project() string {
	return s.boardID
}

// GithubIssue implements source.Issue
type GithubIssue struct {
	title string
	url   string
}

func (i *GithubIssue) Title() string {
	return i.title
}

func (i *GithubIssue) Url() string {
	return i.url
}

// GithubIssueList implements source.IssueList
type GithubIssueList struct {
	issues map[string][]GithubIssue
}

func (i *GithubIssueList) Name(name string) []GithubIssue {
	if list, ok := i.issues[name]; ok {
		return list
	}
	return nil
}

func (s Settings) SourceList() GithubIssueList {
	return s.searchList
}

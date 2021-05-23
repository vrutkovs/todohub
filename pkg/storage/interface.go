package storage

import "github.com/vrutkovs/todohub/pkg/issue"

// Settings holds required methods for source API settings
type Settings interface {
	ID() string
	Authenticate() interface{}
	Project() string
	Self() interface{}
}

// Client holds API
type Client interface {
	CreateProject(string) error
	GetIssues(string) ([]issue.Issue, error)
	Create(string, issue.Issue) error
	Delete(string, issue.Issue) error
}

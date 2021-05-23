package source

import (
	"github.com/vrutkovs/todohub/pkg/issue"
	"github.com/vrutkovs/todohub/pkg/storage"
)

// Settings holds required methods for source API settings
type Settings interface {
	ID() string
	Storage() interface{}
	Project() string
	Searches() map[string]string
	Self() interface{}
}

// Client holds API
type Client interface {
	New(*Settings, *storage.Client)
	Sync() error
	Issues() []issue.Issue
}

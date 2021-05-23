package source

// Settings holds required methods for source API settings
type Settings interface {
	ID() string
	Authenticate() interface{}
	Storage() string
	Project() string
	Lists() SourceList
	Self() interface{}
}

// Client holds API
type Client interface {
	Settings() *Settings
}

// Issue represents an issue in search query
type Issue interface {
	Title() string
	Url() string
}

type IssueList interface {
	Name(string) []Issue
}

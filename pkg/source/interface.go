package source

// Settings holds required methods for source API settings
type Settings interface {
	ID() string
	Authenticate() string
	Storage() string
	Project() string
	Lists() map[string]*Project
	Self() interface{}
}

// Client holds API
type Client interface {
	Settings() *Settings
}

type Project interface {
	Name() string
	Query() string
}

// Issue represents an issue in search query
type Issue interface {
	Title() string
	Url() string
}

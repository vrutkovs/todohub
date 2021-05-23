package storage

// Settings holds required methods for source API settings
type Settings interface {
	ID() string
	Authenticate() interface{}
	Project() string
	Self() interface{}
}

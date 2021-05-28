package todoist

// Settings holds info about trello connnection
type Settings struct {
	Token   string `yaml:"token"`
	Project string `yaml:"project"`
}

// Implement storage.Settings
func (s Settings) ID() string {
	return "todoist"
}

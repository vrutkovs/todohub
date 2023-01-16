package todoist

// Settings holds info about trello connnection
type Settings struct {
	Token       string `yaml:"token"`
	ProjectName string `yaml:"project_name,omitempty"`
	ProjectID   string `yaml:"project_id,omitempty"`
}

// Implement storage.Settings
func (s Settings) ID() string {
	return "todoist"
}

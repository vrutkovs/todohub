package github

// Settings stores info about github connection
type Settings struct {
	Token        string            `yaml:"token"`
	BoardID      string            `yaml:"project,omitempty"`
	SearchPrefix string            `yaml:"search_prefix,omitempty"`
	SearchList   map[string]string `yaml:"lists"`
}

// Implement source.Settings
func (s Settings) ID() string {
	return "github"
}

func (s Settings) Project() string {
	return s.BoardID
}

func (s Settings) Searches() map[string]string {
	return s.SearchList
}

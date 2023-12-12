package jira

type Settings struct {
	Endpoint   string            `yaml:"endpoint"`
	Token      string            `yaml:"token"`
	SearchList map[string]string `yaml:"lists"`
}

// Implement source.Settings.
func (s Settings) ID() string {
	return "jira"
}

func (s Settings) Searches() map[string]string {
	return s.SearchList
}

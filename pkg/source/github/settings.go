package github

// Settings stores info about github connection
type Settings struct {
	token        string            `yaml:"token"`
	storage      string            `yaml:"storage"`
	boardID      string            `yaml:"boardid,omitempty"`
	searchPrefix string            `yaml:"search_prefix,omitempty"`
	searchList   map[string]string `yaml:"lists"`
}

// Implement source.Settings
func (s Settings) ID() string {
	return "github"
}

func (s Settings) Self() interface{} {
	return &s
}

func (s Settings) Storage() interface{} {
	return &s.storage
}

func (s Settings) Project() string {
	return s.boardID
}

func (s Settings) Searches() map[string]string {
	return s.searchList
}

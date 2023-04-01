package trello

// Settings holds info about trello connnection.
type Settings struct {
	AppKey  string `yaml:"appkey"`
	Token   string `yaml:"token"`
	BoardID string `yaml:"boardid"`
}

// Implement storage.Settings.
func (s Settings) ID() string {
	return "trello"
}

func (s Settings) Project() string {
	return s.BoardID
}

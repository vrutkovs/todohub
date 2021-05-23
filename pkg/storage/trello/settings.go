package trello

import api "github.com/adlio/trello"

// Settings holds info about trello connnection
type Settings struct {
	appKey  string `yaml:"appkey"`
	token   string `yaml:"token"`
	boardID string `yaml:"project"`
}

// Implement storage.Settings
func (s Settings) ID() string {
	return "trello"
}

// Pointer to self
func (s Settings) Self() interface{} {
	return s
}

func (s Settings) Authenticate() interface{} {
	c := &Client{
		api:      api.NewClient(s.appKey, s.token),
		settings: &s,
	}
	board, err := c.api.GetBoard(c.settings.boardID, api.Defaults())
	if err != nil {
		panic(err)
	}
	c.board = board
	return c
}

func (s Settings) Project() string {
	return s.boardID
}

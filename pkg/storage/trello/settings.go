package trello

import api "github.com/adlio/trello"

// Settings holds info about trello connnection
type Settings struct {
	AppKey  string `yaml:"appkey"`
	Token   string `yaml:"token"`
	BoardID string `yaml:"boardid"`
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
		api:      api.NewClient(s.AppKey, s.Token),
		settings: &s,
	}
	board, err := c.api.GetBoard(c.settings.BoardID, api.Defaults())
	if err != nil {
		panic(err)
	}
	c.board = board
	return c
}

func (s Settings) Project() string {
	return s.BoardID
}

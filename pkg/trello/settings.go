package trello

// Settings holds info about trello connnection
type Settings struct {
	AppKey  string `yaml:"appkey"`
	Token   string `yaml:"token"`
	BoardID string `yaml:"boardid"`
}

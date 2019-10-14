package trello

type TrelloSettings struct {
	AppKey  string `yaml:"appkey"`
	Token   string `yaml:"token"`
	BoardID string `yaml:"boardid"`
}

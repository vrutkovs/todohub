package settings

import (
	"github.com/vrutkovs/trellohub/pkg/github"
	"github.com/vrutkovs/trellohub/pkg/trello"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const DEFAULT_SYNC_TIMEOUT_MINUTES = 5

type Settings struct {
	Trello      trello.Settings `yaml:"trello"`
	Github      github.Settings `yaml:"github"`
	SyncTimeout uint64          `yaml:"sync_timeout"`
}

func LoadSettings(path string) (*Settings, error) {
	s := Settings{
		SyncTimeout: DEFAULT_SYNC_TIMEOUT_MINUTES,
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

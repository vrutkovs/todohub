package settings

import (
	"io/ioutil"

	"github.com/vrutkovs/todohub/pkg/source/github"
	"github.com/vrutkovs/todohub/pkg/storage"
	"github.com/vrutkovs/todohub/pkg/storage/trello"
	"gopkg.in/yaml.v2"
)

// DefaultSyncTimeoutMinutes sets default sync period
const DefaultSyncTimeoutMinutes = 5

// Settings holds app-level settings
type Settings struct {
	Storage     StorageSettings `yaml:"storage"`
	Source      SourceSettings  `yaml:"source"`
	SyncTimeout uint64          `yaml:"sync_timeout"`
}

// StorageSettings holds storage configs
type StorageSettings struct {
	Trello *trello.Settings `yaml:"trello,omitempty"`
}

// SourceSettings holds client configs
type SourceSettings struct {
	Github *github.Settings `yaml:"github,omitempty"`
}

// LoadSettings creates Settings object from yaml
func LoadSettings(path string) (*Settings, error) {
	s := Settings{
		SyncTimeout: DefaultSyncTimeoutMinutes,
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

func (s *StorageSettings) GetActiveStorage() storage.Settings {
	if s.Trello != nil {
		return s.Trello
	}
	return nil
}

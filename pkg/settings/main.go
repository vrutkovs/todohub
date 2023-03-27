package settings

import (
	"fmt"

	"github.com/vrutkovs/todohub/pkg/source/github"
	"github.com/vrutkovs/todohub/pkg/storage"
	"github.com/vrutkovs/todohub/pkg/storage/todoist"
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
	Trello  *trello.Settings  `yaml:"trello"`
	Todoist *todoist.Settings `yaml:"todoist"`
}

// SourceSettings holds client configs
type SourceSettings struct {
	Github *github.Settings `yaml:"github"`
}

// ReadFile is a function to read file and output a slice of bytes
type ReadFile func(filename string) ([]byte, error)

// LoadSettings creates Settings object from yaml
func LoadSettings(path string, readFile ReadFile) (*Settings, error) {
	s := Settings{
		SyncTimeout: DefaultSyncTimeoutMinutes,
	}

	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *StorageSettings) GetActiveStorageClient() (storage.Client, error) {
	if s.Trello != nil {
		if s.Trello.AppKey != "" && s.Trello.Token != "" && s.Trello.BoardID != "" {
			return trello.New(s.Trello)
		}
	}
	if s.Todoist != nil {
		return todoist.New(s.Todoist)
	}
	return nil, fmt.Errorf("no valid storage settings found")
}

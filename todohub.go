package main

import (
	"io/ioutil"

	"github.com/jasonlvhit/gocron"
	"github.com/vrutkovs/todohub/pkg/settings"
	"github.com/vrutkovs/todohub/pkg/source/github"
	"github.com/vrutkovs/todohub/pkg/source/jira"
)

func main() {
	// Parse settings file
	s, err := settings.LoadSettings("configs/todohub.yaml", ioutil.ReadFile)
	if err != nil {
		panic(err)
	}

	// Find active storage
	storageClient, err := s.Storage.GetActiveStorageClient()
	if err != nil {
		panic(err)
	}

	if s.Source.Github != nil {
		gh := github.New(s.Source.Github, storageClient)
		if err := gocron.Every(s.SyncTimeout).Minutes().Do(gh.Sync, "periodically"); err != nil {
			panic(err)
		}
		if err := gh.Sync("on startup"); err != nil {
			panic(err)
		}
	}

	if s.Source.Jira != nil {
		jiraSource, err := jira.New(s.Source.Jira, storageClient)
		if err != nil {
			panic(err)
		}
		if err := gocron.Every(s.SyncTimeout).Minutes().Do(jiraSource.Sync, "periodically"); err != nil {
			panic(err)
		}
		if err := jiraSource.Sync("on startup"); err != nil {
			panic(err)
		}
	}

	// Start cron
	<-gocron.Start()
}

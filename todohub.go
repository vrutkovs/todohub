package main

import (
	"os"

	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
	"github.com/vrutkovs/todohub/pkg/settings"
	"github.com/vrutkovs/todohub/pkg/source/github"
	"github.com/vrutkovs/todohub/pkg/source/jira"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true, // Enable colors in the output
	})

	// Parse settings file
	s, err := settings.LoadSettings("configs/todohub.yaml", os.ReadFile)
	if err != nil {
		logger.Fatal(err)
	}

	// Find active storage
	storageClient, err := s.Storage.GetActiveStorageClient(logger)
	if err != nil {
		logger.Fatal(err)
	}

	if s.Source.Github != nil {
		gh := github.New(s.Source.Github, storageClient, logger)
		if err := gocron.Every(s.SyncTimeout).Minutes().Do(gh.Sync, "periodically"); err != nil {
			logger.Fatal(err)
		}
		if err := gh.Sync("on startup"); err != nil {
			logger.Fatal(err)
		}
	}

	if s.Source.Jira != nil {
		jiraSource, err := jira.New(s.Source.Jira, storageClient, logger)
		if err != nil {
			logger.Fatal(err)
		}
		if err := gocron.Every(s.SyncTimeout).Minutes().Do(jiraSource.Sync, "periodically"); err != nil {
			logger.Fatal(err)
		}
		if err := jiraSource.Sync("on startup"); err != nil {
			logger.Fatal(err)
		}
	}

	// Start cron
	<-gocron.Start()
}

package main

import (
	"github.com/jasonlvhit/gocron"
	"github.com/vrutkovs/todohub/pkg/settings"
	"github.com/vrutkovs/todohub/pkg/source/github"
)

func main() {
	// Parse settings file
	s, err := settings.LoadSettings("configs/todohub.yaml")
	if err != nil {
		panic(err)
	}

	// Find active storage
	storageClient, err := s.Storage.GetActiveStorageClient()
	if err != nil {
		panic(err)
	}

	if s.Source.Github != nil {
		gh := github.New(s.Source.Github, &storageClient)
		gh.Sync()
		gocron.Every(s.SyncTimeout).Minutes().Do(gh.Sync)
	}

	// Start cron
	<-gocron.Start()

}

package main

import (
	"github.com/jasonlvhit/gocron"
	"github.com/vrutkovs/todohub/pkg/settings"
	"github.com/vrutkovs/todohub/pkg/source/github"
	"io/ioutil"
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
		gh := github.New(s.Source.Github, &storageClient)
		gocron.Every(s.SyncTimeout).Minutes().Do(gh.Sync, "periodically")
		gh.Sync("on startup")
	}

	// Start cron
	<-gocron.Start()

}

package main

import (
	"github.com/jasonlvhit/gocron"
	"github.com/vrutkovs/todohub/pkg/settings"
	"github.com/vrutkovs/todohub/pkg/source/github"
	"github.com/vrutkovs/todohub/pkg/storage/trello"
)

func main() {
	s, err := settings.LoadSettings("configs/todohub.yaml")
	if err != nil {
		panic(err)
	}

	tr := trello.GetClient(s.Trello)
	gh := github.GetClient(s.Github)

	// Update now
	gh.UpdateTrello(tr)

	// Update on cron
	gocron.Every(s.SyncTimeout).Minutes().Do(gh.UpdateTrello, tr)
	<-gocron.Start()

}

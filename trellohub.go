package main

import (
	"github.com/jasonlvhit/gocron"
	"github.com/vrutkovs/trellohub/pkg/github"
	"github.com/vrutkovs/trellohub/pkg/settings"
	"github.com/vrutkovs/trellohub/pkg/trello"
)

func main() {
	s, err := settings.LoadSettings("configs/trellohub.yaml")
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

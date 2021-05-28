module github.com/vrutkovs/todohub

go 1.16

require (
	github.com/adlio/trello v1.9.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/jasonlvhit/gocron v0.0.1
	github.com/kobtea/go-todoist v0.2.2
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/kobtea/go-todoist => github.com/vrutkovs/go-todoist v0.2.3-0.20210529090621-3b6314df3c14

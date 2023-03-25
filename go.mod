module github.com/vrutkovs/todohub

go 1.16

require (
	github.com/adlio/trello v1.10.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/jasonlvhit/gocron v0.0.1
	github.com/kobtea/go-todoist v0.2.2
	golang.org/x/oauth2 v0.4.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/kobtea/go-todoist => github.com/vrutkovs/go-todoist v0.2.3-0.20230325123542-928a5c3dd402

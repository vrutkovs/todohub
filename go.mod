module github.com/vrutkovs/todohub

go 1.16

require (
	github.com/adlio/trello v1.10.0
	github.com/andygrunwald/go-jira/v2 v2.0.0-20231114185916-57d1e28f1bb7
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/jasonlvhit/gocron v0.0.1
	github.com/kobtea/go-todoist v0.2.2
	github.com/onsi/ginkgo/v2 v2.13.2
	github.com/onsi/gomega v1.30.0
	golang.org/x/oauth2 v0.15.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/kobtea/go-todoist => github.com/vrutkovs/go-todoist v0.2.3-0.20230326103331-b1e66b4e3f9a

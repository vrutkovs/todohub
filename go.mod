module github.com/vrutkovs/todohub

go 1.24.0

toolchain go1.24.11

require (
	github.com/adlio/trello v1.12.0
	github.com/andygrunwald/go-jira/v2 v2.0.0-20260113181222-a17356f7cb78
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/jasonlvhit/gocron v0.0.1
	github.com/kobtea/go-todoist v0.2.2
	github.com/onsi/ginkgo/v2 v2.13.2
	github.com/onsi/gomega v1.30.0
	github.com/sirupsen/logrus v1.9.4
	golang.org/x/oauth2 v0.35.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/trivago/tgo v1.0.7 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/kobtea/go-todoist => github.com/vrutkovs/go-todoist v0.2.3-0.20230326103331-b1e66b4e3f9a

package github

type GithubSettings struct {
	Token            string            `yaml:"token"`
	BoardID          string            `yaml:"boardid,omitempty"`
	SearchPrefix     string            `yaml:"search_prefix,omitempty"`
	GithubSearchList map[string]string `yaml:"lists"`
}

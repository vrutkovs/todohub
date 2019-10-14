package github

type GithubSettings struct {
	Token            string            `yaml:"token"`
	SearchPrefix     string            `yaml:"search_prefix,omitempty"`
	GithubSearchList map[string]string `yaml:"lists"`
}

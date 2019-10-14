package github

type GithubSettings struct {
	Token string `yaml:"token"`
	GithubSearchList map[string]string `yaml:"lists"`
}

package github

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGithub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Github")
}

var _ = DescribeTable("repoSlug",
	func(repoUrl, expected string) {
		Expect(repoSlug(repoUrl)).To(Equal(expected))
	},
	Entry("Happy", "https://github.com/vrutkovs/todohub", "vrutkovs/todohub"),
	Entry("Empty", "", ""),
	Entry("Invalid URL", "https://github.com", ""),
)

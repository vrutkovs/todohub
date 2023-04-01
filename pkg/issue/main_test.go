package issue

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIssueInterface(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issues")
}

type IssueMock struct {
	title string
	url   string
	repo  string
}

func (i IssueMock) Title() string {
	return i.title
}

func (i IssueMock) URL() string {
	return i.url
}

func (i IssueMock) Repo() string {
	return i.repo
}

var _ = Describe("Issue List", func() {
	issueA := IssueMock{
		title: "issue A",
		url:   "https://example.com",
		repo:  "vrutkovs/todohub",
	}
	issueB := IssueMock{
		title: "issue B",
		url:   "https://foo.bar",
		repo:  "vrutkovs/example",
	}

	It("can fetch item by name", func() {
		issueList := List{
			Issues: []Issue{issueA, issueB},
		}
		newIssueA, found := issueList.Get(issueA.title)
		Expect(found).To(BeTrue())
		Expect(newIssueA).Should(Equal(issueA))
		newIssueB, found := issueList.Get(issueB.title)
		Expect(found).To(BeTrue())
		Expect(newIssueB).Should(Equal(issueB))
		_, found = issueList.Get("no such issue")
		Expect(found).To(BeFalse())
	})

	It("can remove item by name", func() {
		issueList := List{
			Issues: []Issue{issueA, issueB},
		}
		issueList.Remove(issueB.title)
		Expect(issueList.Issues).Should(Equal([]Issue{issueA}))
		issueList.Remove(issueA.title)
		Expect(issueList.Issues).Should(Equal([]Issue{}))

		issueList = List{
			Issues: []Issue{issueA, issueB},
		}
		issueList.Remove("no-such-issue")
		Expect(issueList.Issues).Should(ContainElements(issueA, issueB))
	})

	It("can build hash list", func() {
		issueList := List{
			Issues: []Issue{issueA, issueB},
		}
		expectedIssueAHash := "7d75692c6fe1a268f891843f82e86df41663178c82cff160890d5cb1a108f7f0"
		expectedIssueBHash := "90d59389ac1f7d881acb43964cceee6e01e2f8dde05647d9ffbc883747137f83"
		Expect(issueList.MakeHashList(false)).Should(Equal(map[string]Issue{
			expectedIssueAHash: issueA,
			expectedIssueBHash: issueB,
		}))
		expectedIssueATitleOnlyHash := "7565421bae35809532c6154bac0777f1f7df2504f5e1bf803d172c70dfa8f8f8"
		expectedIssueBTitleOnlyHash := "1929dc05fcdae8b6ae63c0c7879ce31834eba585f738a2f4b683506e1a9deafd"
		Expect(issueList.MakeHashList(true)).Should(Equal(map[string]Issue{
			expectedIssueATitleOnlyHash: issueA,
			expectedIssueBTitleOnlyHash: issueB,
		}))
	})

	It("can build outer section between two issue lists", func() {
		issueListEmpty := List{
			Issues: []Issue{},
		}
		issueListAll := List{
			Issues: []Issue{issueA, issueB},
		}
		issueListA := List{
			Issues: []Issue{issueA},
		}
		issueListB := List{
			Issues: []Issue{issueB},
		}
		outerSection := OuterSection(issueListAll.MakeHashList(true), issueListEmpty.MakeHashList(true))
		Expect(len(outerSection.Issues)).Should(Equal(2))
		Expect(outerSection.Issues).Should(ContainElements(issueA, issueB))

		outerSection = OuterSection(issueListEmpty.MakeHashList(true), issueListAll.MakeHashList(true))
		Expect(outerSection).Should(Equal(issueListEmpty))

		outerSection = OuterSection(issueListAll.MakeHashList(true), issueListA.MakeHashList(true))
		Expect(outerSection).Should(Equal(issueListB))

		outerSection = OuterSection(issueListAll.MakeHashList(true), issueListB.MakeHashList(true))
		Expect(outerSection).Should(Equal(issueListA))
	})
})

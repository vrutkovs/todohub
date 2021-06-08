package issue

import (
	"reflect"
	"sort"
)

// Issue represents an issue in search query
type Issue interface {
	Title() string
	Url() string
}

type IssueList struct {
	Issues []Issue
}

func (i *IssueList) Get(title string) (Issue, bool) {
	for _, issue := range i.Issues {
		if issue.Title() == title {
			return issue, true
		}
	}
	return nil, false
}

func (i *IssueList) Remove(title string) {
	if _, ok := i.Get(title); !ok {
		return
	}
	newList := make([]Issue, 0)
	for _, issue := range i.Issues {
		if issue.Title() == title {
			continue
		}
		newList = append(newList, issue)
	}
	i.Issues = newList
}

func (a *IssueList) InterSection(b *IssueList, titleOnly bool) IssueList {
	set := IssueList{
		Issues: make([]Issue, 0),
	}

	for _, aEl := range a.Issues {
		idx := sort.Search(len(b.Issues), func(i int) bool {
			bEl := b.Issues[i]
			return compareElements(aEl, bEl, titleOnly)
		})
		if idx < len(b.Issues) && compareElements(b.Issues[idx], aEl, titleOnly) {
			set.Issues = append(set.Issues, aEl)
		}
	}

	return set
}

func compareElements(a, b Issue, titleOnly bool) bool {
	if titleOnly {
		return a.Title() == b.Title()
	} else {
		return reflect.DeepEqual(a, b)
	}
}

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
	newList := make([]Issue, len(i.Issues)-1)
	for _, issue := range i.Issues {
		if issue.Title() == title {
			continue
		}
		newList = append(newList, issue)
	}
	i.Issues = newList
}

func (a *IssueList) InterSection(b IssueList) []interface{} {
	set := make([]interface{}, 0)
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		idx := sort.Search(bv.Len(), func(i int) bool {
			return bv.Index(i).Interface() == el
		})
		if idx < bv.Len() && bv.Index(idx).Interface() == el {
			set = append(set, el)
		}
	}

	return set
}

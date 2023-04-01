package issue

import (
	"crypto/sha256"
	"fmt"
)

// Issue represents an issue in search query.
type Issue interface {
	Title() string
	Url() string
	Repo() string
}

// List represents a list of issues.
type List struct {
	Issues []Issue
}

var h = sha256.New()

func (l *List) Get(title string) (Issue, bool) {
	for _, issue := range l.Issues {
		if issue.Title() == title {
			return issue, true
		}
	}
	return nil, false
}

func (l *List) Remove(title string) {
	if _, ok := l.Get(title); !ok {
		return
	}
	newList := make([]Issue, 0)
	for _, issue := range l.Issues {
		if issue.Title() == title {
			continue
		}
		newList = append(newList, issue)
	}
	l.Issues = newList
}

func asSha256(l Issue, titleOnly bool) string {
	defer h.Reset()
	var obj string
	if titleOnly {
		obj = l.Title()
	} else {
		obj = fmt.Sprintf("%v", l)
	}
	h.Write([]byte(obj))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (l *List) MakeHashList(titleOnly bool) map[string]Issue {
	hashMap := make(map[string]Issue, len(l.Issues))
	for _, aEl := range l.Issues {
		hashMap[asSha256(aEl, titleOnly)] = aEl
	}
	return hashMap
}

func OuterSection(hashA, hashB map[string]Issue) List {
	set := List{
		Issues: make([]Issue, 0),
	}
	for hash, el := range hashA {
		if _, ok := hashB[hash]; !ok {
			set.Issues = append(set.Issues, el)
		}
	}
	return set
}

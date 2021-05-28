package todoist

import (
	"context"
	"fmt"
	"log"
	"regexp"

	api "github.com/kobtea/go-todoist/todoist"
	"github.com/vrutkovs/todohub/pkg/issue"
)

// Client is a wrapper for trello client
type Client struct {
	api      *api.Client
	project  *api.Project
	settings *Settings
	context  *context.Context
}

// Item struct holds information about the card
type Item struct {
	id   api.ID
	text string
}

func (c Item) match() []string {
	re := regexp.MustCompile(`\[(?P<title>.*)\]\((?P<link>.*)\)`)
	m := re.FindAllStringSubmatch(c.text, -1)
	if len(m) == 0 {
		return nil
	}
	return m[0]
}

// Title extracts link title from task contents
func (c Item) Title() string {
	matches := c.match()
	if matches == nil || len(matches) < 3 {
		return ""
	}
	return matches[1]
}

// Url extracts link title from task contents
func (c Item) Url() string {
	matches := c.match()
	if matches == nil || len(matches) < 3 {
		return ""
	}
	return matches[2]
}

// New returns todoist client
func New(s *Settings) *Client {
	clientAPI, err := api.NewClient("", s.Token, "*", "", nil)
	ctx := context.TODO()
	clientAPI.FullSync(ctx, []api.Command{})
	if err != nil {
		panic(err)
	}
	clientAPI.Project.GetAll()
	project := clientAPI.Project.FindOneByName(s.Project)
	if project == nil {
		project, err = api.NewProject(s.Project, &api.NewProjectOpts{})
		if err != nil {
			panic(err)
		}
		_, err := clientAPI.Project.Add(*project)
		if err != nil {
			panic(err)
		}
	}
	return &Client{
		api:      clientAPI,
		project:  project,
		settings: s,
		context:  &ctx,
	}
}

// ensureSectionExists returns list ID if list with this name exists
func (c *Client) ensureSectionExists(name string) (api.ID, error) {
	section := c.api.Section.FindOneByName(name)
	if section != nil {
		return section.ID, nil
	}
	// List was not found, needs to be created
	log.Printf("Creating list %s", name)
	section, err := api.NewSection(name, &api.NewSectionOpts{ParentID: c.project.ID})
	if err != nil {
		return "", err
	}
	_, err = c.api.Section.Add(*section)
	if err != nil {
		return "", err
	}
	return section.ID, nil
}

// CreateProject ensures section is created
func (c *Client) CreateProject(name string) error {
	// c.api.FullSync(*c.context, []api.Command{})
	_, err := c.ensureSectionExists(name)
	return err
}

func apiItemToItem(apiItem *api.Item) Item {
	return Item{
		id:   apiItem.ID,
		text: apiItem.Content,
	}
}

// fetchItemsInSection returns a map of cards
func (c *Client) fetchItemsInSection(sectionID api.ID) ([]Item, error) {
	result := make([]Item, 0)
	allProjectItems := c.api.Item.FindByProjectIDs([]api.ID{c.project.ID})

	for _, item := range allProjectItems {
		if item.SectionID == sectionID {
			result = append(result, apiItemToItem(&item))
		}
	}
	return result, nil
}

func (c *Client) GetIssues(sectionName string) ([]issue.Issue, error) {
	c.api.FullSync(*c.context, []api.Command{})
	issues := make([]issue.Issue, 0)
	sectionID, err := c.ensureSectionExists(sectionName)
	if err != nil {
		return issues, err
	}
	items, err := c.fetchItemsInSection(sectionID)
	if err != nil {
		return issues, err
	}
	// Convert Items back to Issue
	issues = make([]issue.Issue, 0)
	for _, item := range items {
		issues = append(issues, item)
	}
	return issues, nil
}

func buildMarkdownLink(title, url string) string {
	return fmt.Sprintf("[%s](%s)", title, url)
}

func (c *Client) Create(sectionName string, item issue.Issue) error {
	c.api.FullSync(*c.context, []api.Command{})
	sectionID, err := c.ensureSectionExists(sectionName)
	if err != nil {
		return err
	}
	markDownTitle := buildMarkdownLink(item.Title(), item.Url())
	return c.addItemToSection(markDownTitle, sectionID)
}

// addItemToSection adds a text card to the list and return a pointer to Card
func (c *Client) addItemToSection(text string, sectionID api.ID) error {
	item, err := api.NewItem(text, &api.NewItemOpts{
		ProjectID: c.project.ID,
		SectionID: sectionID,
	})
	if err != nil {
		return err
	}
	_, err = c.api.Item.Add(*item)
	if err != nil {
		return err
	}
	return nil
}

// CloseCard marks card as closed and removes it
func (c *Client) Delete(sectionName string, item issue.Issue) error {
	c.api.FullSync(*c.context, []api.Command{})
	// Lookup item by title in the section
	sectionID, err := c.ensureSectionExists(sectionName)
	if err != nil {
		return err
	}
	cardList, err := c.fetchItemsInSection(sectionID)
	if err != nil {
		return err
	}
	for _, i := range cardList {
		if i.Title() == item.Title() {
			err := c.api.Item.Close(i.id)
			if err != nil {
				return err
			}
			break
		}
	}
	return err
}

func (c *Client) Sync() error {
	return c.api.Commit(*c.context)
}

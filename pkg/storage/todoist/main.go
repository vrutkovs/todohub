package todoist

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	api "github.com/kobtea/go-todoist/todoist"
	"github.com/vrutkovs/todohub/pkg/issue"
)

// Client is a wrapper for trello client.
type Client struct {
	api      *api.Client
	project  *api.Project
	settings *Settings
	context  *context.Context
}

// Item struct holds information about the card.
type Item struct {
	id   api.ID
	text string
	repo string
}

var titleRegex = regexp.MustCompile(`\[(?P<title>.*)\]\((?P<link>.*)\)`)

func (c Item) match() []string {
	m := titleRegex.FindAllStringSubmatch(c.text, -1)
	if len(m) == 0 {
		return nil
	}
	return m[0]
}

// Title extracts link title from task contents.
func (c Item) Title() string {
	matches := c.match()
	if matches == nil || len(matches) < 3 {
		return ""
	}
	return matches[1]
}

// URL extracts link title from task contents.
func (c Item) URL() string {
	matches := c.match()
	if matches == nil || len(matches) < 3 {
		return ""
	}
	return matches[2]
}

// Repo extracts valid tags.
func (c Item) Repo() string {
	return c.repo
}

// New returns todoist client.
func New(s *Settings) (*Client, error) {
	clientAPI, err := api.NewClient("", s.Token, "*", "", nil)
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()
	err = clientAPI.FullSync(ctx, []api.Command{})
	if err != nil {
		return nil, err
	}
	var project *api.Project
	if s.ProjectID != "" {
		projectID := api.ID(s.ProjectID)
		projectResponse, err := clientAPI.Project.Get(ctx, projectID)
		if err != nil {
			return nil, err
		}
		project = &projectResponse.Project
	}
	if s.ProjectName != "" {
		project = clientAPI.Project.FindOneByName(s.ProjectName)
	}
	if project == nil {
		project, err = api.NewProject(s.ProjectName, &api.NewProjectOpts{})
		if err != nil {
			return nil, err
		}
		_, err = clientAPI.Project.Add(*project)
		if err != nil {
			return nil, err
		}
	}
	return &Client{
		api:      clientAPI,
		project:  project,
		settings: s,
		context:  &ctx,
	}, nil
}

// ensureSectionExists returns list ID if list with this name exists.
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
	err = c.Sync("after section was created")
	if err != nil {
		return "", err
	}
	return section.ID, nil
}

// ensureLabelExists returns label ID if label with this name exists.
func (c *Client) ensureLabelExists(name string) (string, error) {
	log.Printf("Looking up label %s", name)
	label := c.api.Label.FindOneByName(name)
	if label != nil {
		log.Printf("Found label %#v", label)
		return label.ID, nil
	}
	// Label was not found, needs to be created
	log.Printf("Creating label %s", name)
	label, err := api.NewLabel(name, &api.NewLabelOpts{})
	if err != nil {
		return "", err
	}
	log.Printf("Label to be created %#v", label)
	addedLabel, err := c.api.Label.Add(*label)
	if err != nil {
		return "", err
	}
	log.Printf("Fetched label %#v", addedLabel)
	err = c.Sync("after adding label")
	if err != nil {
		return "", err
	}
	return label.ID, nil
}

// CreateProject ensures section is created.
func (c *Client) CreateProject(name string) error {
	// c.api.FullSync(*c.context, []api.Command{})
	_, err := c.ensureSectionExists(name)
	return err
}

func (c *Client) apiItemToItem(apiItem *api.Item) Item {
	firstLabel := ""
	for _, labelID := range apiItem.Labels {
		label := c.api.Label.FindOneByName(labelID)
		if label == nil {
			continue
		}
		firstLabel = label.Name
		break
	}
	return Item{
		id:   apiItem.ID,
		text: apiItem.Content,
		repo: firstLabel,
	}
}

// fetchItemsInSection returns a map of cards.
func (c *Client) fetchItemsInSection(sectionID api.ID) ([]Item, error) {
	result := make([]Item, 0)
	allProjectItems := c.api.Item.FindByProjectIDs([]api.ID{c.project.ID})

	for _, item := range allProjectItems {
		if item.SectionID == sectionID {
			result = append(result, c.apiItemToItem(&item))
		}
	}
	return result, nil
}

func (c *Client) GetIssues(sectionName string) ([]issue.Issue, error) {
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
	sectionID, err := c.ensureSectionExists(sectionName)
	if err != nil {
		return err
	}
	labelID, err := c.ensureLabelExists(item.Repo())
	if err != nil {
		return err
	}
	markDownTitle := buildMarkdownLink(item.Title(), item.URL())
	return c.addItemToSection(markDownTitle, sectionID, labelID)
}

// addItemToSection adds a text card to the list and return a pointer to Card.
func (c *Client) addItemToSection(text string, sectionID api.ID, labelID string) error {
	item, err := api.NewItem(text, &api.NewItemOpts{
		ProjectID: c.project.ID,
		SectionID: sectionID,
		Labels:    []string{labelID},
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

// CloseCard marks card as closed and removes it.
func (c *Client) Delete(sectionName string, item issue.Issue) error {
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

func (c *Client) Sync(description string) error {
	log.Printf("Syncing %s", description)
	err := c.api.Commit(*c.context)
	if err != nil {
		log.Printf("Error: %s", err)
		time.Sleep(time.Minute * 15)
	}
	err = c.api.FullSync(context.Background(), []api.Command{})
	if err != nil {
		panic(err)
	}
	log.Printf("Done")
	return err
}

// CompareByTitleOnly returns true if issues should be compared by title only
// Some storages may not be able to fetch other details like URL in GetIssues.
func (s *Client) CompareByTitleOnly() bool {
	return true
}

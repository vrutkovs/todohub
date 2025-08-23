package todoist

import (
	"context"
	"fmt"
	"regexp"
	"time"

	api "github.com/kobtea/go-todoist/todoist"
	"github.com/sirupsen/logrus"
	"github.com/vrutkovs/todohub/pkg/issue"
)

// Client is a wrapper for trello client.
type Client struct {
	api      *api.Client
	project  *api.Project
	settings *Settings
	context  *context.Context
	logger   *logrus.Logger
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
	if len(matches) < 3 {
		return ""
	}
	return matches[1]
}

// URL extracts link title from task contents.
func (c Item) URL() string {
	matches := c.match()
	if len(matches) < 3 {
		return ""
	}
	return matches[2]
}

// Repo extracts valid tags.
func (c Item) Repo() string {
	return c.repo
}

// New returns todoist client.
func New(s *Settings, logger *logrus.Logger) (*Client, error) {
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
		logger:   logger,
	}, nil
}

// ensureSectionExists returns list ID if list with this name exists.
func (c *Client) ensureSectionExists(name string) (api.ID, error) {
	logger := c.logger.WithField("storage", "todoist").WithField("section", name)

	logger.Info("looking up section")
	section := c.api.Section.FindOneByName(name)
	if section != nil {
		return section.ID, nil
	}
	// List was not found, needs to be created
	logger.Info("creating section")
	section, err := api.NewSection(name, &api.NewSectionOpts{ParentID: c.project.ID})
	if err != nil {
		logger.WithError(err).Error("failed to compose section")
		return "", err
	}
	_, err = c.api.Section.Add(*section)
	if err != nil {
		logger.WithError(err).Error("failed to add section")
		return "", err
	}
	err = c.Sync("after section was created")
	if err != nil {
		logger.WithError(err).Error("failed to sync after section was created")
		return "", err
	}
	logger.Info("done")
	return section.ID, nil
}

// ensureLabelExists returns label ID if label with this name exists.
func (c *Client) ensureLabelExists(name string) (string, error) {
	logger := c.logger.WithField("storage", "todoist").WithField("label", name)
	logger.Info("looking up label")
	label := c.api.Label.FindOneByName(name)
	if label != nil {
		return label.Name, nil
	}
	// Label was not found, needs to be created
	logger.Info("creating label")
	label, err := api.NewLabel(name, &api.NewLabelOpts{})
	if err != nil {
		logger.WithError(err).Error("failed to create label")
		return "", err
	}
	_, err = c.api.Label.Add(*label)
	if err != nil {
		logger.WithError(err).Error("failed to add label")
		return "", err
	}
	err = c.Sync("after adding label")
	if err != nil {
		logger.WithError(err).Error("failed to sync after adding label")
		return "", err
	}
	logger.Info("done")
	return label.Name, nil
}

// CreateProject ensures section is created.
func (c *Client) CreateProject(name string) error {
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
func (c *Client) fetchItemsInSection(sectionID api.ID, sectionName string) []Item {
	logger := c.logger.WithField("storage", "todoist").WithField("section", sectionName)
	logger.Info("fetching items")
	result := make([]Item, 0)
	allProjectItems := c.api.Item.FindByProjectIDs([]api.ID{c.project.ID})

	for i, item := range allProjectItems {
		if item.SectionID == sectionID {
			result = append(result, c.apiItemToItem(&allProjectItems[i]))
		}
	}
	logger.WithField("count", len(result)).Info("fetched items")
	return result
}

func (c *Client) GetIssues(sectionName string) ([]issue.Issue, error) {
	issues := make([]issue.Issue, 0)
	sectionID, err := c.ensureSectionExists(sectionName)
	if err != nil {
		return issues, err
	}
	items := c.fetchItemsInSection(sectionID, sectionName)
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
	labelName, err := c.ensureLabelExists(item.Repo())
	if err != nil {
		return err
	}
	markDownTitle := buildMarkdownLink(item.Title(), item.URL())
	return c.addItemToSection(markDownTitle, sectionID, sectionName, labelName)
}

// addItemToSection adds a text card to the list and return a pointer to Card.
func (c *Client) addItemToSection(text string, sectionID api.ID, sectionName, labelName string) error {
	logger := c.logger.WithField("storage", "todoist").WithField("section", sectionName).WithField("text", text)
	logger.Info("adding item")

	item, err := api.NewItem(text, &api.NewItemOpts{
		ProjectID: c.project.ID,
		SectionID: sectionID,
		Labels:    []string{labelName},
	})
	if err != nil {
		logger.WithError(err).Error("failed to create item")
		return err
	}
	_, err = c.api.Item.Add(*item)
	if err != nil {
		logger.WithError(err).Error("failed to add item")
		return err
	}
	return nil
}

// CloseCard marks card as closed and removes it.
func (c *Client) Delete(sectionName string, item issue.Issue) error {
	logger := c.logger.WithField("storage", "todoist").WithField("section", sectionName).WithField("item", item.Title())
	logger.Info("deleting item")

	// Lookup item by title in the section
	sectionID, err := c.ensureSectionExists(sectionName)
	if err != nil {
		logger.WithError(err).Error("failed to ensure section exists")
		return err
	}
	cardList := c.fetchItemsInSection(sectionID, sectionName)
	for _, i := range cardList {
		if i.Title() == item.Title() {
			err := c.api.Item.Close(i.id)
			if err != nil {
				logger.WithError(err).Error("failed to close item")
				return err
			}
			break
		}
	}
	return nil
}

func (c *Client) Sync(description string) error {
	logger := c.logger.WithField("storage", "todoist").WithField("description", description)
	logger.Info("syncing")
	err := c.api.Commit(*c.context)
	if err != nil {
		logger.WithError(err).Error("failed to commit")
		time.Sleep(time.Minute * 15)
	}
	err = c.api.FullSync(context.Background(), []api.Command{})
	if err != nil {
		c.logger.Fatal(err)
	}
	logger.Info("done")
	return err
}

// CompareByTitleOnly returns true if issues should be compared by title only
// Some storages may not be able to fetch other details like URL in GetIssues.
func (c *Client) CompareByTitleOnly() bool {
	return true
}

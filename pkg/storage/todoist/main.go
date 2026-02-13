package todoist

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/gofrs/uuid"
	todoist "github.com/sachaos/todoist/lib"
	"github.com/sirupsen/logrus"
	"github.com/vrutkovs/todohub/pkg/issue"
)

// Client is a wrapper for todoist client.
type Client struct {
	api      *todoist.Client
	project  *todoist.Project
	settings *Settings
	logger   *logrus.Logger
}

// Item struct holds information about the card.
type Item struct {
	id   string
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
	config := &todoist.Config{
		AccessToken: s.Token,
	}
	client := todoist.NewClient(config)
	ctx := context.Background()
	if err := client.Sync(ctx); err != nil {
		return nil, err
	}

	var project *todoist.Project
	if s.ProjectID != "" {
		for _, p := range client.Store.Projects {
			if p.ID == s.ProjectID {
				project = &p
				break
			}
		}
	}
	if project == nil && s.ProjectName != "" {
		for _, p := range client.Store.Projects {
			if p.Name == s.ProjectName {
				project = &p
				break
			}
		}
	}

	if project == nil {
		// Create project
		id, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		newProj := todoist.Project{
			Name: s.ProjectName,
		}
		newProj.ID = id.String()

		if err := client.AddProject(ctx, newProj); err != nil {
			return nil, err
		}
		if err := client.Sync(ctx); err != nil {
			return nil, err
		}
		for _, p := range client.Store.Projects {
			if p.Name == s.ProjectName {
				project = &p
				break
			}
		}
	}

	if project == nil {
		return nil, fmt.Errorf("failed to find or create project")
	}

	return &Client{
		api:      client,
		project:  project,
		settings: s,
		logger:   logger,
	}, nil
}

// ensureSectionExists returns list ID if list with this name exists.
func (c *Client) ensureSectionExists(name string) (string, error) {
	logger := c.logger.WithField("storage", "todoist").WithField("section", name)

	logger.Info("looking up section")
	for _, s := range c.api.Store.Sections {
		if s.ProjectID == c.project.ID && s.Name == name {
			return s.ID, nil
		}
	}

	// List was not found, needs to be created
	logger.Info("creating section")
	tmpID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	args := map[string]interface{}{
		"name":       name,
		"project_id": c.project.ID,
	}
	cmd := todoist.NewCommand("section_add", args)
	cmd.TempID = tmpID.String()

	if err := c.api.ExecCommands(context.Background(), todoist.Commands{cmd}); err != nil {
		logger.WithError(err).Error("failed to add section")
		return "", err
	}

	if err := c.Sync("after section was created"); err != nil {
		logger.WithError(err).Error("failed to sync after section was created")
		return "", err
	}

	// Find again
	for _, s := range c.api.Store.Sections {
		if s.ProjectID == c.project.ID && s.Name == name {
			logger.Info("done")
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("failed to find section after creation")
}

// ensureLabelExists returns label ID if label with this name exists.
func (c *Client) ensureLabelExists(name string) (string, error) {
	logger := c.logger.WithField("storage", "todoist").WithField("label", name)
	logger.Info("looking up label")

	for _, l := range c.api.Store.Labels {
		if l.Name == name {
			return l.ID, nil
		}
	}

	// Label was not found, needs to be created
	logger.Info("creating label")
	tmpID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	args := map[string]interface{}{
		"name": name,
	}
	cmd := todoist.NewCommand("label_add", args)
	cmd.TempID = tmpID.String()

	if err := c.api.ExecCommands(context.Background(), todoist.Commands{cmd}); err != nil {
		logger.WithError(err).Error("failed to add label")
		return "", err
	}

	if err := c.Sync("after adding label"); err != nil {
		logger.WithError(err).Error("failed to sync after adding label")
		return "", err
	}

	for _, l := range c.api.Store.Labels {
		if l.Name == name {
			logger.Info("done")
			return l.ID, nil
		}
	}
	return "", fmt.Errorf("failed to find label after creation")
}

// CreateProject ensures section is created.
func (c *Client) CreateProject(name string) error {
	_, err := c.ensureSectionExists(name)
	return err
}

func (c *Client) apiItemToItem(apiItem todoist.Item) Item {
	firstLabel := ""
	if len(apiItem.LabelNames) > 0 {
		labelID := apiItem.LabelNames[0]
		if label := c.api.Store.FindLabel(labelID); label != nil {
			firstLabel = label.Name
		}
	}
	return Item{
		id:   apiItem.ID,
		text: apiItem.Content,
		repo: firstLabel,
	}
}

// fetchItemsInSection returns a map of cards.
func (c *Client) fetchItemsInSection(sectionID, sectionName string) []Item {
	logger := c.logger.WithField("storage", "todoist").WithField("section", sectionName)
	logger.Info("fetching items")
	result := make([]Item, 0)

	for _, item := range c.api.Store.Items {
		if item.ProjectID == c.project.ID && item.SectionID == sectionID {
			result = append(result, c.apiItemToItem(item))
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
	labelID, err := c.ensureLabelExists(item.Repo())
	if err != nil {
		return err
	}
	markDownTitle := buildMarkdownLink(item.Title(), item.URL())
	return c.addItemToSection(markDownTitle, sectionID, sectionName, labelID)
}

// addItemToSection adds a text card to the list and return a pointer to Card.
func (c *Client) addItemToSection(text, sectionID, sectionName, labelID string) error {
	logger := c.logger.WithField("storage", "todoist").WithField("section", sectionName).WithField("text", text)
	logger.Info("adding item")

	item := todoist.Item{}
	item.Content = text
	item.ProjectID = c.project.ID
	item.SectionID = sectionID
	item.LabelNames = []string{labelID}

	if err := c.api.AddItem(context.Background(), item); err != nil {
		logger.WithError(err).Error("failed to create item")
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
			err := c.api.CloseItem(context.Background(), []string{i.id})
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
	err := c.api.Sync(context.Background())
	if err != nil {
		logger.WithError(err).Error("failed to sync")
		time.Sleep(time.Minute * 15)
	}
	logger.Info("done")
	return err
}

// CompareByTitleOnly returns true if issues should be compared by title only
// Some storages may not be able to fetch other details like URL in GetIssues.
func (c *Client) CompareByTitleOnly() bool {
	return true
}

package trello

import (
	"log"

	api "github.com/adlio/trello"
	"github.com/vrutkovs/todohub/pkg/issue"
)

// Client is a wrapper for trello client.
type Client struct {
	api      *api.Client
	board    *api.Board
	settings *Settings
}

// Card struct holds information about the card.
type Card struct {
	id    string
	title string
	url   string
}

func (c Card) Title() string {
	return c.title
}

func (c Card) URL() string {
	// TODO: Fetch first attached URL
	return c.url
}

func (c Card) Repo() string {
	// TODO: Find repo in contents
	return ""
}

// New returns trello client.
func New(s *Settings) (*Client, error) {
	clientAPI := api.NewClient(s.AppKey, s.Token)
	board, err := clientAPI.GetBoard(s.BoardID, api.Defaults())
	if err != nil {
		return nil, err
	}
	return &Client{
		api:      clientAPI,
		board:    board,
		settings: s,
	}, nil
}

// ensureListExists returns list ID if list with this name exists.
func (c *Client) ensureListExists(name string) (string, error) {
	lists, err := c.board.GetLists(api.Defaults())
	if err != nil {
		return "", err
	}
	for _, list := range lists {
		if list.Name == name {
			return list.ID, nil
		}
	}
	// List was not found, needs to be created
	log.Printf("Creating list %s", name)
	list, err := c.api.CreateList(c.board, name, api.Defaults())
	if err != nil {
		return "", err
	}
	return list.ID, nil
}

func (c *Client) CreateProject(name string) error {
	_, err := c.ensureListExists(name)
	return err
}

func apiCardToCard(apiCard *api.Card) Card {
	var url string
	if len(apiCard.Attachments) == 0 {
		url = ""
	} else {
		url = apiCard.Attachments[0].URL
	}
	return Card{
		id:    apiCard.ID,
		title: apiCard.Name,
		url:   url,
	}
}

// FetchCardsInList returns a map of cards.
func (c *Client) fetchCardsInList(listID string) ([]Card, error) {
	list, err := c.api.GetList(listID, api.Defaults())
	if err != nil {
		return nil, err
	}
	// Check that the card doesn't exist yet
	apiCards, err := list.GetCards(api.Arguments{"filter": "all"})
	if err != nil {
		return nil, err
	}

	result := make([]Card, 0)
	for _, apiCard := range apiCards {
		if apiCard.Closed {
			continue
		}
		result = append(result, apiCardToCard(apiCard))
	}

	return result, err
}

func (c *Client) GetIssues(listName string) ([]issue.Issue, error) {
	issues := make([]issue.Issue, 0)
	listID, err := c.ensureListExists(listName)
	if err != nil {
		return issues, err
	}
	cards, err := c.fetchCardsInList(listID)
	if err != nil {
		return issues, err
	}
	// Convert Cards back to Issue
	issues = make([]issue.Issue, 0)
	for _, card := range cards {
		issues = append(issues, card)
	}
	return issues, nil
}

func (c *Client) Create(listName string, item issue.Issue) error {
	listID, err := c.ensureListExists(listName)
	if err != nil {
		return err
	}
	card, err := c.addItemToList(item.Title(), listID)
	if err != nil {
		return err
	}
	return c.attachLink(card, item.URL())
}

// AddItemToList adds a text card to the list and return a pointer to Card.
func (c *Client) addItemToList(item, listID string) (*Card, error) {
	list, err := c.api.GetList(listID, api.Defaults())
	if err != nil {
		return nil, err
	}
	// Check that the card doesn't exist yet
	cards, err := list.GetCards(api.Arguments{"filter": "all"})
	if err != nil {
		return nil, err
	}
	for _, card := range cards {
		if card.Name == item {
			return &Card{
				id:    card.ID,
				title: card.Name,
			}, nil
		}
	}
	// Create a new card
	apiCard := &api.Card{Name: item}
	err = list.AddCard(apiCard, api.Defaults())
	if err != nil {
		return nil, err
	}
	if apiCard.Closed {
		if err := apiCard.Unarchive(); err != nil {
			return nil, err
		}
	}
	return &Card{
		id:    apiCard.ID,
		title: apiCard.Name,
	}, nil
}

// AttachLink adds a URL as attachment to the card.
func (c *Client) attachLink(card *Card, url string) error {
	apiCard, err := c.api.GetCard(card.id, api.Arguments{"attachments": "true"})
	if err != nil {
		return err
	}
	for _, attach := range apiCard.Attachments {
		if attach.URL == url {
			return nil
		}
	}
	attach := &api.Attachment{URL: url}
	if err := apiCard.AddURLAttachment(attach); err != nil {
		return err
	}
	log.Printf("trello: attached url %s", url)
	return nil
}

// CloseCard marks card as closed and removes it.
func (c *Client) Delete(listName string, item issue.Issue) error {
	// Lookup card by title in the list
	listID, err := c.ensureListExists(listName)
	if err != nil {
		return err
	}
	cardList, err := c.fetchCardsInList(listID)
	if err != nil {
		return err
	}
	for _, i := range cardList {
		if i.Title() != item.Title() {
			continue
		}
		// Mark card as closed
		card, err := c.api.GetCard(i.id, api.Defaults())
		if err != nil {
			return err
		}
		if card.Closed {
			return nil
		}
		err = card.Archive()
		if err != nil {
			return err
		}
		err = card.Delete()
		if err != nil {
			return err
		}
	}
	return err
}

// Sync ensures changes are committed.
func (c *Client) Sync(_ string) error {
	// trello changes are auto-synced
	return nil
}

// CompareByTitleOnly returns true if issues should be compared by title only
// Some storages may not be able to fetch other details like URL in GetIssues.
func (c *Client) CompareByTitleOnly() bool {
	return true
}

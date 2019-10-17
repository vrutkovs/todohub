package trello

import (
	api "github.com/adlio/trello"
	"log"
)

type Client struct {
	api      *api.Client
	board    *api.Board
	settings TrelloSettings
}

type Card struct {
	id    string
	title string
}

func GetClient(settings TrelloSettings) *Client {
	c := &Client{
		api:      api.NewClient(settings.AppKey, settings.Token),
		settings: settings,
	}
	board, err := c.api.GetBoard(c.settings.BoardID, api.Defaults())
	if err != nil {
		panic(err)
	}
	c.board = board
	return c
}

func (c *Client) SetBoardID(id string) {
	log.Printf("trello: using board %s", id)
	board, err := c.api.GetBoard(id, api.Defaults())
	if err != nil {
		panic(err)
	}
	c.board = board
}

func (c *Client) EnsureListExists(name string) (string, error) {
	log.Printf("Creating list %s", name)
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
	list, err := c.api.CreateList(c.board, name, api.Defaults())
	if err != nil {
		return "", err
	}
	return list.ID, nil
}

func (c *Client) AddItemToList(item string, listID string) (*Card, error) {
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
	return &Card{
		id:    apiCard.ID,
		title: apiCard.Name,
	}, nil
}

func (c *Client) AttachLink(card *Card, url string) error {
	apiCard, err := c.api.GetCard(card.id, api.Arguments{"attachments": "true"})
	if err != nil {
		return err
	}
	for _, attach := range apiCard.Attachments {
		if attach.URL == url {
			log.Printf("URL %s is already attached to this card", url)
			return nil
		}
	}
	attach := &api.Attachment{URL: url}
	apiCard.AddURLAttachment(attach)
	log.Printf("github: attached url %s", url)
	return nil
}

func (c *Client) FetchCardsInList(listID string) (map[string]string, error) {
	list, err := c.api.GetList(listID, api.Defaults())
	if err != nil {
		return nil, err
	}
	// Check that the card doesn't exist yet
	cards, err := list.GetCards(api.Arguments{"filter": "all"})
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, 0)
	for _, card := range cards {
		result[card.Name] = card.ID
	}

	return result, err
}

func (c *Client) CloseCard(id string) error {
	card, err := c.api.GetCard(id, api.Defaults())
	if err != nil {
		return err
	}
	if card.Closed == true {
		return nil
	}
	card.Closed = true
	card.Update(api.Defaults())
	log.Printf("Card %s closed", card.Name)
	return nil
}

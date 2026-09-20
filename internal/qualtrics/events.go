package qualtrics

import (
	"context"
	"encoding/json"
)

type EventSubscription struct {
	ID              string `json:"id"`
	Scope           string `json:"scope"`
	Topics          string `json:"topics"`
	PublicationURL  string `json:"publicationUrl"`
	Encrypt         bool   `json:"encrypt"`
	SuccessfulCalls int    `json:"successfulCalls"`
}

func (c *Client) CreateEventSubscription(ctx context.Context, payload json.RawMessage) (EventSubscription, error) {
	var out EventSubscription
	if err := c.Do(ctx, "POST", "/eventsubscriptions", payload, &out); err != nil {
		return EventSubscription{}, err
	}
	return out, nil
}

func (c *Client) GetEventSubscription(ctx context.Context, subscriptionID string) (EventSubscription, error) {
	var out EventSubscription
	if err := c.Do(ctx, "GET", "/eventsubscriptions/"+subscriptionID, nil, &out); err != nil {
		return EventSubscription{}, err
	}
	return out, nil
}

func (c *Client) DeleteEventSubscription(ctx context.Context, subscriptionID string) error {
	return c.Do(ctx, "DELETE", "/eventsubscriptions/"+subscriptionID, nil, nil)
}

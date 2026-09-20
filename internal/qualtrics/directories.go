package qualtrics

import (
	"context"
	"encoding/json"
)

type Directory struct {
	DirectoryID string `json:"directoryId"`
	Name        string `json:"name"`
}

type directoryListResponse struct {
	Elements []Directory `json:"elements"`
	NextPage string      `json:"nextPage"`
}

func (c *Client) ListDirectories(ctx context.Context, opts ListOptions) (Page[Directory], error) {
	path := pathWithQuery("/directories", opts)
	var out directoryListResponse
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Page[Directory]{}, err
	}
	next := -1
	if out.NextPage != "" {
		next = opts.Offset + len(out.Elements)
	}
	return NewPage(out.Elements, len(out.Elements), next), nil
}

func (c *Client) GetDirectory(ctx context.Context, directoryID string) (Directory, error) {
	var out Directory
	if err := c.Do(ctx, "GET", "/directories/"+directoryID, nil, &out); err != nil {
		return Directory{}, err
	}
	return out, nil
}

func (c *Client) CreateDirectory(ctx context.Context, payload json.RawMessage) (Directory, error) {
	var out Directory
	if err := c.Do(ctx, "POST", "/directories", payload, &out); err != nil {
		return Directory{}, err
	}
	return out, nil
}

func (c *Client) UpdateDirectory(ctx context.Context, directoryID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/directories/"+directoryID, payload, nil)
}

func (c *Client) DeleteDirectory(ctx context.Context, directoryID string) error {
	return c.Do(ctx, "DELETE", "/directories/"+directoryID, nil, nil)
}

type MailingList struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DirectoryID string `json:"directoryId"`
}

type mailingListResponse struct {
	Elements []MailingList `json:"elements"`
	NextPage string        `json:"nextPage"`
}

func (c *Client) ListMailingLists(ctx context.Context, directoryID string, opts ListOptions) (Page[MailingList], error) {
	path := pathWithQuery("/directories/"+directoryID+"/mailinglists", opts)
	var out mailingListResponse
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Page[MailingList]{}, err
	}
	next := -1
	if out.NextPage != "" {
		next = opts.Offset + len(out.Elements)
	}
	return NewPage(out.Elements, len(out.Elements), next), nil
}

func (c *Client) GetMailingList(ctx context.Context, directoryID, mailingListID string) (MailingList, error) {
	var out MailingList
	if err := c.Do(ctx, "GET", "/directories/"+directoryID+"/mailinglists/"+mailingListID, nil, &out); err != nil {
		return MailingList{}, err
	}
	return out, nil
}

func (c *Client) CreateMailingList(ctx context.Context, directoryID string, payload json.RawMessage) (MailingList, error) {
	var out MailingList
	if err := c.Do(ctx, "POST", "/directories/"+directoryID+"/mailinglists", payload, &out); err != nil {
		return MailingList{}, err
	}
	return out, nil
}

func (c *Client) UpdateMailingList(ctx context.Context, directoryID, mailingListID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/directories/"+directoryID+"/mailinglists/"+mailingListID, payload, nil)
}

func (c *Client) DeleteMailingList(ctx context.Context, directoryID, mailingListID string) error {
	return c.Do(ctx, "DELETE", "/directories/"+directoryID+"/mailinglists/"+mailingListID, nil, nil)
}

type Contact struct {
	ContactID    string         `json:"contactId"`
	FirstName    string         `json:"firstName"`
	LastName     string         `json:"lastName"`
	Email        string         `json:"email"`
	Phone        string         `json:"phone"`
	Unsubscribed bool           `json:"unsubscribed"`
	EmbeddedData map[string]any `json:"embeddedData"`
}

type contactListResponse struct {
	Elements          []Contact `json:"elements"`
	NextPage          string    `json:"nextPage"`
	ContinuationToken string    `json:"continuationToken"`
}

func (c *Client) ListContacts(ctx context.Context, directoryID, mailingListID string, opts ListOptions) (Page[Contact], error) {
	path := pathWithQuery("/directories/"+directoryID+"/mailinglists/"+mailingListID+"/contacts", opts)
	var out contactListResponse
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Page[Contact]{}, err
	}
	next := -1
	if out.NextPage != "" || out.ContinuationToken != "" {
		next = opts.Offset + len(out.Elements)
	}
	return NewPage(out.Elements, len(out.Elements), next), nil
}

func (c *Client) GetContact(ctx context.Context, directoryID, mailingListID, contactID string) (Contact, error) {
	var out Contact
	path := "/directories/" + directoryID + "/mailinglists/" + mailingListID + "/contacts/" + contactID
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Contact{}, err
	}
	return out, nil
}

func (c *Client) CreateContact(ctx context.Context, directoryID, mailingListID string, payload json.RawMessage) (Contact, error) {
	var out Contact
	path := "/directories/" + directoryID + "/mailinglists/" + mailingListID + "/contacts"
	if err := c.Do(ctx, "POST", path, payload, &out); err != nil {
		return Contact{}, err
	}
	return out, nil
}

func (c *Client) UpdateContact(ctx context.Context, directoryID, mailingListID, contactID string, payload json.RawMessage) error {
	path := "/directories/" + directoryID + "/mailinglists/" + mailingListID + "/contacts/" + contactID
	return c.Do(ctx, "PUT", path, payload, nil)
}

func (c *Client) DeleteContact(ctx context.Context, directoryID, mailingListID, contactID string) error {
	path := "/directories/" + directoryID + "/mailinglists/" + mailingListID + "/contacts/" + contactID
	return c.Do(ctx, "DELETE", path, nil, nil)
}

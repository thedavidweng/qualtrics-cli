package qualtrics

import (
	"context"
	"encoding/json"
	"net/url"
)

type Distribution struct {
	ID            string `json:"id"`
	SurveyID      string `json:"surveyId"`
	RequestStatus string `json:"requestStatus"`
	SendDate      string `json:"sendDate"`
	CreatedDate   string `json:"createdDate"`
	ModifiedDate  string `json:"modifiedDate"`
	Type          string `json:"type"`
	Description   string `json:"description"`
}

type distributionListResponse struct {
	Elements []Distribution `json:"elements"`
	NextPage string         `json:"nextPage"`
}

func (c *Client) ListDistributions(ctx context.Context, surveyID string, opts ListOptions) (Page[Distribution], error) {
	q := opts.Query()
	if surveyID != "" {
		q.Set("surveyId", surveyID)
	}
	path := "/distributions"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out distributionListResponse
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Page[Distribution]{}, err
	}
	next := -1
	if out.NextPage != "" {
		next = opts.Offset + len(out.Elements)
	}
	return NewPage(out.Elements, len(out.Elements), next), nil
}

func (c *Client) GetDistribution(ctx context.Context, distributionID, surveyID string) (Distribution, error) {
	path := "/distributions/" + distributionID
	if surveyID != "" {
		path += "?surveyId=" + url.QueryEscape(surveyID)
	}
	var out Distribution
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Distribution{}, err
	}
	return out, nil
}

func (c *Client) CreateDistribution(ctx context.Context, payload json.RawMessage) (Distribution, error) {
	var out Distribution
	if err := c.Do(ctx, "POST", "/distributions", payload, &out); err != nil {
		return Distribution{}, err
	}
	return out, nil
}

func (c *Client) DeleteDistribution(ctx context.Context, distributionID string) error {
	return c.Do(ctx, "DELETE", "/distributions/"+distributionID, nil, nil)
}

type DistributionLink struct {
	LinkID          string `json:"linkId"`
	TransactionID   string `json:"transactionId"`
	Link            string `json:"link"`
	Status          string `json:"status"`
	LastName        string `json:"lastName"`
	FirstName       string `json:"firstName"`
	Email           string `json:"email"`
	ExternalDataRef string `json:"externalDataReference"`
}

type distributionLinkListResponse struct {
	Elements []DistributionLink `json:"elements"`
	NextPage string             `json:"nextPage"`
}

func (c *Client) ListDistributionLinks(ctx context.Context, distributionID, surveyID string, opts ListOptions) (Page[DistributionLink], error) {
	q := opts.Query()
	if surveyID != "" {
		q.Set("surveyId", surveyID)
	}
	path := "/distributions/" + distributionID + "/links"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out distributionLinkListResponse
	if err := c.Do(ctx, "GET", path, nil, &out); err != nil {
		return Page[DistributionLink]{}, err
	}
	next := -1
	if out.NextPage != "" {
		next = opts.Offset + len(out.Elements)
	}
	return NewPage(out.Elements, len(out.Elements), next), nil
}

func (c *Client) GetDistributionLink(ctx context.Context, distributionID, linkID string) (DistributionLink, error) {
	var out DistributionLink
	if err := c.Do(ctx, "GET", "/distributions/"+distributionID+"/links/"+linkID, nil, &out); err != nil {
		return DistributionLink{}, err
	}
	return out, nil
}

func (c *Client) CreateDistributionLinks(ctx context.Context, distributionID string, payload json.RawMessage) (any, error) {
	var out any
	if err := c.Do(ctx, "POST", "/distributions/"+distributionID+"/links", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateDistributionLink(ctx context.Context, distributionID, linkID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/distributions/"+distributionID+"/links/"+linkID, payload, nil)
}

func (c *Client) DeleteDistributionLink(ctx context.Context, distributionID, linkID string) error {
	return c.Do(ctx, "DELETE", "/distributions/"+distributionID+"/links/"+linkID, nil, nil)
}

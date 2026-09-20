package qualtrics

import (
	"context"
)

type Survey struct {
	SurveyID     string `json:"SurveyID"`
	SurveyName   string `json:"SurveyName"`
	OwnerID      string `json:"OwnerID"`
	BrandID      string `json:"BrandID"`
	DivisionID   string `json:"DivisionID"`
	SurveyStatus string `json:"SurveyStatus"`
	CreationDate string `json:"CreationDate"`
	LastModified string `json:"LastModified"`
	LastAccessed string `json:"LastAccessed"`
}

type SurveyList struct {
	Elements []Survey `json:"elements"`
}

func (c *Client) ListSurveys(ctx context.Context, opts ListOptions) (Page[Survey], error) {
	var out SurveyList
	if err := c.Do(ctx, "GET", pathWithQuery("/surveys", opts), nil, &out); err != nil {
		return Page[Survey]{}, err
	}
	next := -1
	if len(out.Elements) > 0 {
		next = opts.Offset + len(out.Elements)
	}
	return NewPage(out.Elements, len(out.Elements), next), nil
}

func (c *Client) GetSurvey(ctx context.Context, surveyID string) (Survey, error) {
	var out Survey
	if err := c.Do(ctx, "GET", "/surveys/"+surveyID, nil, &out); err != nil {
		return Survey{}, err
	}
	return out, nil
}

type CreateSurveyRequest struct {
	SurveyName      string `json:"SurveyName"`
	Language        string `json:"Language"`
	ProjectCategory string `json:"ProjectCategory"`
}

func (c *Client) CreateSurvey(ctx context.Context, req CreateSurveyRequest) (Survey, error) {
	var out Survey
	if err := c.Do(ctx, "POST", "/survey-definitions", req, &out); err != nil {
		return Survey{}, err
	}
	return out, nil
}

func (c *Client) DeleteSurvey(ctx context.Context, surveyID string) error {
	return c.Do(ctx, "DELETE", "/survey-definitions/"+surveyID, nil, nil)
}

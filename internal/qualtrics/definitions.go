package qualtrics

import (
	"context"
	"encoding/json"
)

type BlockElement struct {
	Type       string `json:"Type"`
	QuestionID string `json:"QuestionID"`
}

type Block struct {
	ID            string         `json:"ID"`
	Type          string         `json:"Type"`
	Description   string         `json:"Description"`
	BlockElements []BlockElement `json:"BlockElements"`
}

type Definition struct {
	SurveyID     string                     `json:"SurveyID"`
	SurveyName   string                     `json:"SurveyName"`
	Blocks       map[string]Block           `json:"Blocks"`
	Questions    map[string]json.RawMessage `json:"Questions"`
	Flow         json.RawMessage            `json:"Flow"`
	SurveyOption json.RawMessage            `json:"SurveyOption"`
}

func (d *Definition) BlockList() []Block {
	blocks := make([]Block, 0, len(d.Blocks))
	for _, b := range d.Blocks {
		blocks = append(blocks, b)
	}
	return blocks
}

func (d *Definition) Summary() map[string]any {
	perBlock := make([]map[string]any, 0, len(d.Blocks))
	questions := 0
	for _, b := range d.Blocks {
		n := 0
		for _, el := range b.BlockElements {
			if el.Type == "Question" {
				n++
			}
		}
		questions += n
		perBlock = append(perBlock, map[string]any{
			"id":          b.ID,
			"description": b.Description,
			"questions":   n,
		})
	}
	return map[string]any{
		"survey_id":   d.SurveyID,
		"survey_name": d.SurveyName,
		"blocks":      len(d.Blocks),
		"questions":   questions,
		"per_block":   perBlock,
	}
}

func (c *Client) GetDefinition(ctx context.Context, surveyID string) (Definition, error) {
	var out Definition
	if err := c.Do(ctx, "GET", "/survey-definitions/"+surveyID, nil, &out); err != nil {
		return Definition{}, err
	}
	return out, nil
}

func (c *Client) CreateFromDefinition(ctx context.Context, definition json.RawMessage) (Definition, error) {
	var out Definition
	if err := c.Do(ctx, "POST", "/survey-definitions", definition, &out); err != nil {
		return Definition{}, err
	}
	return out, nil
}

type Question struct {
	QuestionID          string          `json:"QuestionID"`
	QuestionText        string          `json:"QuestionText"`
	QuestionType        string          `json:"QuestionType"`
	Selector            string          `json:"Selector"`
	SubSelector         string          `json:"SubSelector"`
	DataExportTag       string          `json:"DataExportTag"`
	QuestionDescription string          `json:"QuestionDescription"`
	Choices             json.RawMessage `json:"Choices"`
	Configuration       json.RawMessage `json:"Configuration"`
}

func decodeQuestions(raw json.RawMessage) ([]Question, error) {
	var list []Question
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	var byID map[string]Question
	if err := json.Unmarshal(raw, &byID); err == nil {
		out := make([]Question, 0, len(byID))
		for id := range byID {
			q := byID[id]
			if q.QuestionID == "" {
				q.QuestionID = id
			}
			out = append(out, q)
		}
		return out, nil
	}
	return nil, nil
}

func (c *Client) ListQuestions(ctx context.Context, surveyID string) ([]Question, error) {
	var raw json.RawMessage
	if err := c.Do(ctx, "GET", "/survey-definitions/"+surveyID+"/questions", nil, &raw); err != nil {
		return nil, err
	}
	return decodeQuestions(raw)
}

func (c *Client) GetQuestion(ctx context.Context, surveyID, questionID string) (Question, error) {
	var out Question
	if err := c.Do(ctx, "GET", "/survey-definitions/"+surveyID+"/questions/"+questionID, nil, &out); err != nil {
		return Question{}, err
	}
	return out, nil
}

func (c *Client) CreateQuestion(ctx context.Context, surveyID, blockID string, payload json.RawMessage) (Question, error) {
	path := "/survey-definitions/" + surveyID + "/questions"
	if blockID != "" {
		path += "?blockId=" + blockID
	}
	var out Question
	if err := c.Do(ctx, "POST", path, payload, &out); err != nil {
		return Question{}, err
	}
	return out, nil
}

func (c *Client) UpdateQuestion(ctx context.Context, surveyID, questionID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/survey-definitions/"+surveyID+"/questions/"+questionID, payload, nil)
}

func (c *Client) DeleteQuestion(ctx context.Context, surveyID, questionID string) error {
	return c.Do(ctx, "DELETE", "/survey-definitions/"+surveyID+"/questions/"+questionID, nil, nil)
}

func (c *Client) GetBlock(ctx context.Context, surveyID, blockID string) (Block, error) {
	var out Block
	if err := c.Do(ctx, "GET", "/survey-definitions/"+surveyID+"/blocks/"+blockID, nil, &out); err != nil {
		return Block{}, err
	}
	return out, nil
}

type CreateBlockRequest struct {
	Description string `json:"Description"`
	Type        string `json:"Type"`
}

func (c *Client) CreateBlock(ctx context.Context, surveyID, description string) (Block, error) {
	var out Block
	req := CreateBlockRequest{Description: description, Type: "Standard"}
	if err := c.Do(ctx, "POST", "/survey-definitions/"+surveyID+"/blocks", req, &out); err != nil {
		return Block{}, err
	}
	return out, nil
}

func (c *Client) UpdateBlock(ctx context.Context, surveyID, blockID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/survey-definitions/"+surveyID+"/blocks/"+blockID, payload, nil)
}

func (c *Client) DeleteBlock(ctx context.Context, surveyID, blockID string) error {
	return c.Do(ctx, "DELETE", "/survey-definitions/"+surveyID+"/blocks/"+blockID, nil, nil)
}

func (c *Client) GetFlow(ctx context.Context, surveyID string) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.Do(ctx, "GET", "/survey-definitions/"+surveyID+"/flow", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateFlow(ctx context.Context, surveyID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/survey-definitions/"+surveyID+"/flow", payload, nil)
}

func (c *Client) GetOptions(ctx context.Context, surveyID string) (json.RawMessage, error) {
	var out json.RawMessage
	if err := c.Do(ctx, "GET", "/survey-definitions/"+surveyID+"/options", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateOptions(ctx context.Context, surveyID string, payload json.RawMessage) error {
	return c.Do(ctx, "PUT", "/survey-definitions/"+surveyID+"/options", payload, nil)
}

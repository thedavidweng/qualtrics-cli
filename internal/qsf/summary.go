package qsf

import (
	"encoding/json"
	"fmt"
)

type SummaryInfo struct {
	SurveyID    string         `json:"survey_id"`
	SurveyName  string         `json:"survey_name"`
	Language    string         `json:"language"`
	BlockCount  int            `json:"block_count"`
	TotalQCount int            `json:"total_question_count"`
	Blocks      []BlockSummary `json:"blocks"`
	Elements    map[string]int `json:"element_counts"`
}

type BlockSummary struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Questions   int    `json:"questions"`
}

func SummarizeQSF(data []byte) (*SummaryInfo, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse QSF JSON: %w", err)
	}

	entry, _ := raw["SurveyEntry"].(map[string]any)
	surveyID, _ := entry["SurveyID"].(string)
	surveyName, _ := entry["SurveyName"].(string)
	lang, _ := entry["SurveyLanguage"].(string)

	elements, _ := raw["SurveyElements"].([]any)
	elementCounts := make(map[string]int)

	var blocks []BlockSummary
	totalQ := 0

	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		elType, _ := m["Element"].(string)
		elementCounts[elType]++

		if elType == "BL" {
			blPayload, ok := m["Payload"].([]any)
			if ok {
				for _, b := range blPayload {
					bm, ok := b.(map[string]any)
					if !ok {
						continue
					}
					bType, _ := bm["Type"].(string)
					if bType == "Trash" {
						continue
					}
					bID, _ := bm["ID"].(string)
					bDesc, _ := bm["Description"].(string)
					elems, _ := bm["BlockElements"].([]any)
					qCount := 0
					for _, item := range elems {
						im, ok := item.(map[string]any)
						if ok && im["Type"] == "Question" {
							qCount++
						}
					}
					blocks = append(blocks, BlockSummary{
						ID:          bID,
						Description: bDesc,
						Type:        bType,
						Questions:   qCount,
					})
				}
			}
		}

		if elType == "SQ" {
			totalQ++
		}
	}

	return &SummaryInfo{
		SurveyID:    surveyID,
		SurveyName:  surveyName,
		Language:    lang,
		BlockCount:  len(blocks),
		TotalQCount: totalQ,
		Blocks:      blocks,
		Elements:    elementCounts,
	}, nil
}

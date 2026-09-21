package qsf

import (
	"encoding/json"
	"fmt"
)

type DefinitionOutput struct {
	SurveyID     string                     `json:"SurveyID"`
	SurveyName   string                     `json:"SurveyName"`
	Blocks       map[string]any             `json:"Blocks"`
	Questions    map[string]json.RawMessage `json:"Questions"`
	Flow         json.RawMessage            `json:"Flow"`
	SurveyOption json.RawMessage            `json:"SurveyOption"`
}

func ConvertQSFToDefinition(data []byte) ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse QSF JSON: %w", err)
	}

	entry, _ := raw["SurveyEntry"].(map[string]any)
	surveyID, _ := entry["SurveyID"].(string)
	surveyName, _ := entry["SurveyName"].(string)

	elements, _ := raw["SurveyElements"].([]any)

	blocksMap := make(map[string]any)
	questionsMap := make(map[string]json.RawMessage)
	var flowRaw json.RawMessage
	var optionsRaw json.RawMessage

	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			continue
		}
		elType, _ := m["Element"].(string)

		switch elType {
		case "BL":
			blPayload, ok := m["Payload"].([]any)
			if ok {
				for _, b := range blPayload {
					bm, ok := b.(map[string]any)
					if !ok {
						continue
					}
					if bm["Type"] == "Trash" {
						continue
					}
					bID, _ := bm["ID"].(string)
					if bID != "" {
						blocksMap[bID] = bm
					}
				}
			}
		case "FL":
			b, _ := json.Marshal(m["Payload"])
			flowRaw = b
		case "SO":
			b, _ := json.Marshal(m["Payload"])
			optionsRaw = b
		case "SQ":
			qid, _ := m["PrimaryAttribute"].(string)
			if qid != "" {
				b, _ := json.Marshal(m["Payload"])
				questionsMap[qid] = b
			}
		}
	}

	def := DefinitionOutput{
		SurveyID:     surveyID,
		SurveyName:   surveyName,
		Blocks:       blocksMap,
		Questions:    questionsMap,
		Flow:         flowRaw,
		SurveyOption: optionsRaw,
	}

	return json.MarshalIndent(def, "", "  ")
}

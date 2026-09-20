package qsf

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

type typeInfo struct {
	qt  string
	sel string
	sub string
}

var questionTypes = map[string]typeInfo{
	"mc":           {"MC", "SAVR", "TX"},
	"mc-multi":     {"MC", "MAVR", "TX"},
	"mc-dropdown":  {"MC", "DL", "TX"},
	"rank":         {"RO", "DND", "TX"},
	"text":         {"TE", "SL", ""},
	"text-essay":   {"TE", "ESTB", ""},
	"matrix":       {"Matrix", "Likert", "SingleAnswer"},
	"matrix-multi": {"Matrix", "Likert", "MultipleAnswer"},
	"description":  {"DB", "TB", ""},
}

const alphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randID(prefix string) string {
	const length = 15
	b := make([]byte, length)
	for i := range b {
		b[i] = alphaNum[rand.IntN(len(alphaNum))]
	}
	return prefix + string(b)
}

func joinBody(lines []string) string {
	var paragraphs []string
	var current []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			current = append(current, trimmed)
		} else if len(current) > 0 {
			paragraphs = append(paragraphs, strings.Join(current, " "))
			current = nil
		}
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, strings.Join(current, " "))
	}
	return strings.Join(paragraphs, "<br><br>")
}

func logicConditionBlock(dl *LogicCondition) map[string]any {
	locator := fmt.Sprintf("q://%s/SelectableChoice/%d", dl.QID, dl.Choice)
	opText := "Is " + dl.Operator
	if dl.Operator == "Selected" {
		opText = "Is Selected"
	}
	desc := fmt.Sprintf(`<span class="ConjDesc">If</span> <span class="QuestionDesc">%s</span> <span class="LeftOpDesc">Choice %d</span> <span class="OpDesc">%s</span> `, dl.QID, dl.Choice, opText)

	return map[string]any{
		"0": map[string]any{
			"0": map[string]any{
				"LogicType":             "Question",
				"QuestionID":            dl.QID,
				"QuestionIsInLoop":      "no",
				"ChoiceLocator":         locator,
				"Operator":              dl.Operator,
				"QuestionIDFromLocator": dl.QID,
				"LeftOperand":           locator,
				"Type":                  "Expression",
				"Description":           desc,
			},
			"Type": "If",
		},
		"Type": "BooleanExpression",
	}
}

func buildDisplayLogic(dl *LogicCondition) map[string]any {
	block := logicConditionBlock(dl)
	block["inPage"] = false
	return block
}

func buildBranchLogic(dl *LogicCondition) map[string]any {
	return logicConditionBlock(dl)
}

func buildSkipLogicEntry(sl *SkipRule, qid string, q *QuestionSpec, skipID int) map[string]any {
	var locator string
	var choiceDisplay string
	if q.Type == "text" || q.Type == "text-essay" {
		locator = fmt.Sprintf("q://%s/ChoiceTextEntryValue", qid)
		choiceDisplay = q.Text
		if len(choiceDisplay) > 40 {
			choiceDisplay = choiceDisplay[:40]
		}
	} else {
		locator = fmt.Sprintf("q://%s/SelectableChoice/%d", qid, sl.Choice)
		if sl.Choice > 0 && sl.Choice <= len(q.Choices) {
			choiceDisplay = q.Choices[sl.Choice-1]
		} else {
			choiceDisplay = fmt.Sprintf("Choice %d", sl.Choice)
		}
	}

	destLabels := map[string]string{
		"ENDOFSURVEY": "End of Survey",
		"ENDOFBLOCK":  "End of Block",
	}
	destDesc := sl.Destination
	if d, ok := destLabels[sl.Destination]; ok {
		destDesc = d
	}

	condLabels := map[string]string{
		"Selected":    "Is Selected",
		"NotSelected": "Is Not Selected",
		"Empty":       "Is Empty",
		"NotEmpty":    "Is Not Empty",
	}
	condDesc := sl.Condition
	if c, ok := condLabels[sl.Condition]; ok {
		condDesc = c
	}

	qTextShort := q.Text
	if len(qTextShort) > 40 {
		qTextShort = qTextShort[:40]
	}

	return map[string]any{
		"SkipLogicID":       skipID,
		"ChoiceLocator":     locator,
		"Condition":         sl.Condition,
		"SkipToDestination": sl.Destination,
		"Locator":           locator,
		"SkipToDescription": fmt.Sprintf("%s <strong>%s</strong>  <strong>%s</strong>", qTextShort, choiceDisplay, condDesc),
		"Description":       fmt.Sprintf(`Condition: <strong title=%q>%s</strong> <strong>%s</strong>. Skip To: <strong>%s</strong>.`, choiceDisplay, choiceDisplay, condDesc, destDesc),
		"QuestionID":        qid,
	}
}

func buildLanguageDict(q *QuestionSpec) map[string]any {
	allLangs := make(map[string]bool)
	for l := range q.TextTranslations {
		allLangs[l] = true
	}
	for l := range q.ScaleTranslations {
		allLangs[l] = true
	}
	for l := range q.ChoiceTranslations {
		allLangs[l] = true
	}
	for l := range q.RowTranslations {
		allLangs[l] = true
	}

	result := make(map[string]any)
	for lang := range allLangs {
		entry := make(map[string]any)
		if t, ok := q.TextTranslations[lang]; ok {
			entry["QuestionText"] = t
		}
		if strings.HasPrefix(q.Type, "matrix") {
			if rowTrans, ok := q.RowTranslations[lang]; ok && len(rowTrans) > 0 {
				choicesMap := make(map[string]any)
				for i := 0; i < len(q.Rows); i++ {
					if t, ok := rowTrans[i+1]; ok {
						choicesMap[strconv.Itoa(i+1)] = map[string]any{"Display": t}
					}
				}
				entry["Choices"] = choicesMap
			}
			if scaleTrans, ok := q.ScaleTranslations[lang]; ok && len(scaleTrans) > 0 {
				answersMap := make(map[string]any)
				for i, s := range scaleTrans {
					answersMap[strconv.Itoa(i+1)] = map[string]any{"Display": s}
				}
				entry["Answers"] = answersMap
			}
		} else {
			if choiceTrans, ok := q.ChoiceTranslations[lang]; ok && len(choiceTrans) > 0 {
				choicesMap := make(map[string]any)
				for i := 0; i < len(q.Choices); i++ {
					if t, ok := choiceTrans[i+1]; ok {
						choicesMap[strconv.Itoa(i+1)] = map[string]any{"Display": t}
					}
				}
				entry["Choices"] = choicesMap
			}
		}
		if len(entry) > 0 {
			result[lang] = entry
		}
	}
	return result
}

func buildQuestionPayload(q *QuestionSpec, qidNum int, blockDisplayLogic *LogicCondition) map[string]any {
	qid := fmt.Sprintf("QID%d", qidNum)
	tag := q.Label
	if tag == "" {
		tag = fmt.Sprintf("Q%d", qidNum)
	}

	ti, ok := questionTypes[q.Type]
	if !ok {
		ti = questionTypes["text"]
	}

	force := "OFF"
	if q.Required {
		force = "ON"
	}

	qText := q.Text
	if ti.qt == "DB" && len(q.BodyLines) > 0 {
		qText = joinBody(q.BodyLines)
	}

	qDesc := q.Text
	if len(qDesc) > 99 {
		qDesc = qDesc[:99]
	}

	langDict := buildLanguageDict(q)
	var langVal any = langDict
	if len(langDict) == 0 {
		langVal = []any{}
	}

	payload := map[string]any{
		"QuestionText":        qText,
		"DataExportTag":       tag,
		"QuestionType":        ti.qt,
		"Selector":            ti.sel,
		"DataVisibility":      map[string]any{"Private": false, "Hidden": false},
		"Configuration":       map[string]any{"QuestionDescriptionOption": "UseText"},
		"QuestionDescription": qDesc,
		"Validation": map[string]any{
			"Settings": map[string]any{
				"ForceResponse": force,
				"Type":          "None",
			},
		},
		"Language":     langVal,
		"NextChoiceId": 1,
		"NextAnswerId": 1,
	}

	if ti.sub != "" {
		payload["SubSelector"] = ti.sub
	}

	if ti.qt == "MC" || ti.qt == "RO" {
		if q.CarryFrom != "" {
			payload["Choices"] = []any{}
			payload["ChoiceOrder"] = []any{}
			payload["NextChoiceId"] = 1
			payload["DynamicChoices"] = map[string]any{
				"DynamicType": "ChoiceGroup",
				"Locator":     fmt.Sprintf("q://%s/ChoiceGroup/SelectedChoices", q.CarryFrom),
				"Type":        "Dynamic",
			}
			payload["DynamicChoicesData"] = []any{}
		} else {
			choiceDict := make(map[string]any)
			choiceOrder := make([]string, len(q.Choices))
			for i, c := range q.Choices {
				key := strconv.Itoa(i + 1)
				entry := map[string]any{"Display": c}
				if q.TextEntryChoices[i+1] {
					entry["TextEntry"] = "true"
				}
				choiceDict[key] = entry
				choiceOrder[i] = key
			}
			payload["Choices"] = choiceDict
			payload["ChoiceOrder"] = choiceOrder
			payload["NextChoiceId"] = len(q.Choices) + 1
			payload["DynamicChoicesData"] = []any{}
		}
	}

	if ti.qt == "TE" {
		payload["SearchSource"] = map[string]any{"AllowFreeResponse": "false"}
	}

	if ti.qt == "Matrix" {
		choices := make(map[string]any)
		choiceOrder := make([]string, len(q.Rows))
		for i, r := range q.Rows {
			key := strconv.Itoa(i + 1)
			choices[key] = map[string]any{"Display": r}
			choiceOrder[i] = key
		}

		answers := make(map[string]any)
		answerOrder := make([]string, len(q.Scale))
		for i, s := range q.Scale {
			key := strconv.Itoa(i + 1)
			answers[key] = map[string]any{"Display": s}
			answerOrder[i] = key
		}

		payload["DefaultChoices"] = false
		payload["Choices"] = choices
		payload["ChoiceOrder"] = choiceOrder
		payload["Answers"] = answers
		payload["AnswerOrder"] = answerOrder
		payload["NextChoiceId"] = len(q.Rows) + 1
		payload["NextAnswerId"] = len(q.Scale) + 1
		payload["GradingData"] = []any{}
		payload["ChoiceDataExportTags"] = false
		payload["Configuration"] = map[string]any{
			"QuestionDescriptionOption": "UseText",
			"TextPosition":              "inline",
			"ChoiceColumnWidth":         40,
			"RepeatHeaders":             "none",
			"WhiteSpace":                "OFF",
			"MobileFirst":               true,
		}
	}

	effectiveDL := q.DisplayLogic
	if effectiveDL == nil {
		effectiveDL = blockDisplayLogic
	}
	if effectiveDL != nil {
		payload["DisplayLogic"] = buildDisplayLogic(effectiveDL)
	}

	payload["QuestionID"] = qid

	if len(q.RecodeValues) > 0 {
		recodes := make(map[string]string)
		for k, v := range q.RecodeValues {
			recodes[strconv.Itoa(k)] = v
		}
		payload["RecodeValues"] = recodes
	}
	if len(q.VariableNames) > 0 {
		varnames := make(map[string]string)
		for k, v := range q.VariableNames {
			varnames[strconv.Itoa(k)] = v
		}
		payload["VariableNaming"] = varnames
	}

	return payload
}

type questionTriple struct {
	qidNum            int
	question          *QuestionSpec
	blockDisplayLogic *LogicCondition
}

type blockData struct {
	id          string
	name        string
	blockType   string
	elements    []map[string]any
	branchLogic *LogicCondition
	loopFrom    string
}

func BuildQSF(survey *SurveySpec) (map[string]any, error) {
	surveyID := randID("SV_")
	userID := randID("UR_")
	rsID := randID("RS_")
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	qidCounter := 1
	skipIDCounter := 1
	blocksData := make([]*blockData, 0, len(survey.Blocks))
	var allTriples []questionTriple

	for i, block := range survey.Blocks {
		blockID := randID("BL_")
		var blockElements []map[string]any
		blockDL := block.DisplayLogic

		for _, q := range block.Questions {
			if q.IsPageBreak {
				blockElements = append(blockElements, map[string]any{"Type": "Page Break"})
			} else {
				qid := fmt.Sprintf("QID%d", qidCounter)
				elem := map[string]any{"Type": "Question", "QuestionID": qid}
				if len(q.SkipLogic) > 0 {
					var skipEntries []map[string]any
					for k, sl := range q.SkipLogic {
						skipEntries = append(skipEntries, buildSkipLogicEntry(sl, qid, q, skipIDCounter+k))
					}
					skipIDCounter += len(q.SkipLogic)
					elem["SkipLogic"] = skipEntries
				}
				blockElements = append(blockElements, elem)
				allTriples = append(allTriples, questionTriple{
					qidNum:            qidCounter,
					question:          q,
					blockDisplayLogic: blockDL,
				})
				qidCounter++
			}
		}

		bType := "Standard"
		if i == 0 {
			bType = "Default"
		}

		blocksData = append(blocksData, &blockData{
			id:          blockID,
			name:        block.Name,
			blockType:   bType,
			elements:    blockElements,
			branchLogic: block.BranchLogic,
			loopFrom:    block.LoopFrom,
		})
	}

	labelToQID := make(map[string]string)
	for _, t := range allTriples {
		if t.question.Label != "" {
			labelToQID[t.question.Label] = fmt.Sprintf("QID%d", t.qidNum)
		}
	}

	resolve := func(ref string) string {
		if strings.HasPrefix(ref, "@") {
			label := ref[1:]
			if r, ok := labelToQID[label]; ok {
				return r
			}
		}
		return ref
	}

	for _, t := range allTriples {
		q := t.question
		if q.CarryFrom != "" {
			q.CarryFrom = resolve(q.CarryFrom)
		}
		if q.DisplayLogic != nil {
			q.DisplayLogic.QID = resolve(q.DisplayLogic.QID)
		}
		for _, sl := range q.SkipLogic {
			sl.Destination = resolve(sl.Destination)
		}
	}

	for _, bd := range blocksData {
		if bd.loopFrom != "" {
			bd.loopFrom = resolve(bd.loopFrom)
		}
		if bd.branchLogic != nil {
			bd.branchLogic.QID = resolve(bd.branchLogic.QID)
		}
	}

	qidToQuestion := make(map[string]*QuestionSpec)
	for _, t := range allTriples {
		qidToQuestion[fmt.Sprintf("QID%d", t.qidNum)] = t.question
	}

	blPayload := make([]any, 0, len(blocksData)+1)
	for _, bd := range blocksData {
		entry := map[string]any{
			"Type":          bd.blockType,
			"Description":   bd.name,
			"ID":            bd.id,
			"BlockElements": bd.elements,
		}
		if bd.loopFrom != "" {
			srcQ := qidToQuestion[bd.loopFrom]
			nChoices := 0
			if srcQ != nil {
				nChoices = len(srcQ.Choices)
			}
			locator := fmt.Sprintf("q://%s/ChoiceGroup/SelectedChoices", bd.loopFrom)
			staticMap := make(map[string]any)
			for j := 0; j < nChoices; j++ {
				staticMap[strconv.Itoa(j+1)] = map[string]any{"2": ""}
			}
			entry["SubType"] = ""
			entry["Options"] = map[string]any{
				"BlockLocking":       "false",
				"RandomizeQuestions": "false",
				"Looping":            "Question",
				"LoopingOptions": map[string]any{
					"Locator":            locator,
					"QID":                bd.loopFrom,
					"ChoiceGroupLocator": locator,
					"Static":             staticMap,
					"Randomization":      "None",
				},
			}
		}
		blPayload = append(blPayload, entry)
	}

	blPayload = append(blPayload, map[string]any{
		"Type":        "Trash",
		"Description": "Trash / Unused Questions",
		"ID":          randID("BL_"),
	})

	blElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "BL",
		"PrimaryAttribute":   "Survey Blocks",
		"SecondaryAttribute": nil,
		"TertiaryAttribute":  nil,
		"Payload":            blPayload,
	}

	flowCounter := 2
	var flowItems []any
	for _, bd := range blocksData {
		switch {
		case bd.loopFrom != "":
			flowItems = append(flowItems, map[string]any{
				"ID":     bd.id,
				"Type":   "Standard",
				"FlowID": fmt.Sprintf("FL_%d", flowCounter),
			})
			flowCounter++
		case bd.branchLogic != nil:
			branchFlowID := fmt.Sprintf("FL_%d", flowCounter)
			flowCounter++
			nestedFlowID := fmt.Sprintf("FL_%d", flowCounter)
			flowCounter++
			flowItems = append(flowItems, map[string]any{
				"Type":        "Branch",
				"FlowID":      branchFlowID,
				"Description": "New Branch",
				"BranchLogic": buildBranchLogic(bd.branchLogic),
				"Flow": []any{
					map[string]any{
						"Type":     "Block",
						"ID":       bd.id,
						"FlowID":   nestedFlowID,
						"Autofill": []any{},
					},
				},
			})
		default:
			flowItems = append(flowItems, map[string]any{
				"ID":       bd.id,
				"Type":     "Block",
				"FlowID":   fmt.Sprintf("FL_%d", flowCounter),
				"Autofill": []any{},
			})
			flowCounter++
		}
	}

	flElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "FL",
		"PrimaryAttribute":   "Survey Flow",
		"SecondaryAttribute": nil,
		"TertiaryAttribute":  nil,
		"Payload": map[string]any{
			"Flow":       flowItems,
			"Properties": map[string]any{"Count": flowCounter - 1},
			"FlowID":     "FL_1",
			"Type":       "Root",
		},
	}

	projElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "PROJ",
		"PrimaryAttribute":   "CORE",
		"SecondaryAttribute": nil,
		"TertiaryAttribute":  "1.1.0",
		"Payload":            map[string]any{"ProjectCategory": "CORE", "SchemaVersion": "1.1.0"},
	}

	qcElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "QC",
		"PrimaryAttribute":   "Survey Question Count",
		"SecondaryAttribute": strconv.Itoa(len(allTriples)),
		"TertiaryAttribute":  nil,
		"Payload":            nil,
	}

	rsElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "RS",
		"PrimaryAttribute":   rsID,
		"SecondaryAttribute": "Default Response Set",
		"TertiaryAttribute":  nil,
		"Payload":            nil,
	}

	scoElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "SCO",
		"PrimaryAttribute":   "Scoring",
		"SecondaryAttribute": nil,
		"TertiaryAttribute":  nil,
		"Payload": map[string]any{
			"ScoringCategories":            []any{},
			"ScoringCategoryGroups":        []any{},
			"ScoringSummaryCategory":       nil,
			"ScoringSummaryAfterQuestions": 0,
			"ScoringSummaryAfterSurvey":    0,
			"DefaultScoringCategory":       nil,
			"AutoScoringCategory":          nil,
		},
	}

	soPayload := map[string]any{
		"BackButton":                  "false",
		"SaveAndContinue":             "true",
		"SurveyProtection":            "PublicSurvey",
		"BallotBoxStuffingPrevention": "false",
		"NoIndex":                     "Yes",
		"SecureResponseFiles":         "true",
		"SurveyExpiration":            "None",
		"SurveyTermination":           "DefaultMessage",
		"Header":                      "",
		"Footer":                      "",
		"ProgressBarDisplay":          "None",
		"PartialData":                 "+1 week",
		"ValidationMessage":           "",
		"PreviousButton":              "",
		"NextButton":                  "",
		"SurveyTitle":                 survey.Title,
		"SkinLibrary":                 "qualtrics",
		"SkinType":                    "templated",
		"Skin":                        map[string]any{"brandingId": nil, "templateId": "*base", "overrides": nil},
		"NewScoring":                  1,
		"SurveyMetaDescription":       "",
	}

	if len(survey.ExtraLanguages) > 0 {
		avail := map[string]any{survey.Language: []any{}}
		for _, el := range survey.ExtraLanguages {
			avail[el] = []any{}
		}
		soPayload["AvailableLanguages"] = avail
		soPayload["HiddenLanguages"] = []any{}
		soPayload["CustomLanguages"] = []any{}
		soPayload["MetaDataTranslations"] = map[string]any{}
	}

	soElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "SO",
		"PrimaryAttribute":   "Survey Options",
		"SecondaryAttribute": nil,
		"TertiaryAttribute":  nil,
		"Payload":            soPayload,
	}

	sqElements := make([]any, 0, len(allTriples))
	for _, t := range allTriples {
		qDesc := t.question.Text
		if len(qDesc) > 99 {
			qDesc = qDesc[:99]
		}
		sqElements = append(sqElements, map[string]any{
			"SurveyID":           surveyID,
			"Element":            "SQ",
			"PrimaryAttribute":   fmt.Sprintf("QID%d", t.qidNum),
			"SecondaryAttribute": qDesc,
			"TertiaryAttribute":  nil,
			"Payload":            buildQuestionPayload(t.question, t.qidNum, t.blockDisplayLogic),
		})
	}

	statElement := map[string]any{
		"SurveyID":           surveyID,
		"Element":            "STAT",
		"PrimaryAttribute":   "Survey Statistics",
		"SecondaryAttribute": nil,
		"TertiaryAttribute":  nil,
		"Payload":            map[string]any{"MobileCompatible": true, "ID": "Survey Statistics"},
	}

	var surveyDesc any = survey.Description
	if survey.Description == "" {
		surveyDesc = nil
	}

	elements := make([]any, 0, 8+len(sqElements))
	elements = append(elements,
		blElement,
		flElement,
		projElement,
		qcElement,
		rsElement,
		scoElement,
		soElement,
	)
	elements = append(elements, sqElements...)
	elements = append(elements, statElement)

	return map[string]any{
		"SurveyEntry": map[string]any{
			"SurveyID":                surveyID,
			"SurveyName":              survey.Title,
			"SurveyDescription":       surveyDesc,
			"SurveyOwnerID":           userID,
			"SurveyBrandID":           "qualtrics",
			"DivisionID":              nil,
			"SurveyLanguage":          survey.Language,
			"SurveyActiveResponseSet": rsID,
			"SurveyStatus":            "Inactive",
			"SurveyStartDate":         "0000-00-00 00:00:00",
			"SurveyExpirationDate":    "0000-00-00 00:00:00",
			"SurveyCreationDate":      now,
			"CreatorID":               userID,
			"LastModified":            now,
			"LastAccessed":            "0000-00-00 00:00:00",
			"LastActivated":           "0000-00-00 00:00:00",
			"Deleted":                 nil,
		},
		"SurveyElements": elements,
	}, nil
}

var _ = sort.Strings

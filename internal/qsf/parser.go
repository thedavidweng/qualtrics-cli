package qsf

import (
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	reQuestion         = regexp.MustCompile(`^##\s+(.+?)\s+\[([^\]]+)\]\s*(\*)?\s*(?:@(\w+))?\s*$`)
	reLogicCondition   = regexp.MustCompile(`(?i)(@\w+|QID\d+)/(\d+)\s+(\w+)`)
	reSkipRule         = regexp.MustCompile(`(?i)(\d+)\s+(\w+)\s+[→>]\s+(\S+)`)
	reLoopFrom         = regexp.MustCompile(`(?i)(@\w+|QID\d+)`)
	reChoiceAnnotation = regexp.MustCompile(`^(.*?)\s*\[([A-Z_][A-Z0-9_]*)(?:=(\d+))?\]\s*$`)
	reTextEntry        = regexp.MustCompile(`(?i)\s*\[\+text\]\s*`)
	reLangLine         = regexp.MustCompile(`(?i)^lang-([a-z]{2})(?:-(scale))?:\s*(.+)$`)
	reBlock            = regexp.MustCompile(`^#\s+`)
)

type itemKind int

const (
	kindNone itemKind = iota
	kindQuestionText
	kindScale
	kindChoice
	kindRow
)

type trackedItem struct {
	kind  itemKind
	index int
}

func parseLogicCondition(line string) *LogicCondition {
	m := reLogicCondition.FindStringSubmatch(line)
	if len(m) < 4 {
		return nil
	}
	raw := m[1]
	qid := raw
	if !strings.HasPrefix(raw, "@") {
		qid = strings.ToUpper(raw)
	}
	choice, _ := strconv.Atoi(m[2])
	return &LogicCondition{
		QID:      qid,
		Choice:   choice,
		Operator: m[3],
	}
}

func parseSkipRule(line string) *SkipRule {
	m := reSkipRule.FindStringSubmatch(line)
	if len(m) < 4 {
		return nil
	}
	choice, _ := strconv.Atoi(m[1])
	return &SkipRule{
		Choice:      choice,
		Condition:   m[2],
		Destination: m[3],
	}
}

func newEmptyQuestion(text, qtype string, required bool) *QuestionSpec {
	return &QuestionSpec{
		Text:               text,
		Type:               strings.ToLower(qtype),
		Required:           required,
		Choices:            []string{},
		Rows:               []string{},
		Scale:              []string{},
		BodyLines:          []string{},
		SkipLogic:          []*SkipRule{},
		RecodeValues:       make(map[int]string),
		VariableNames:      make(map[int]string),
		TextEntryChoices:   make(map[int]bool),
		TextTranslations:   make(map[string]string),
		ChoiceTranslations: make(map[string]map[int]string),
		RowTranslations:    make(map[string]map[int]string),
		ScaleTranslations:  make(map[string][]string),
	}
}

func ParseSurvey(text string) (*SurveySpec, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	idx := 0

	meta := make(map[string]any)
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		idx = 1
		var fm []string
		for idx < len(lines) && strings.TrimSpace(lines[idx]) != "---" {
			fm = append(fm, lines[idx])
			idx++
		}
		if idx < len(lines) {
			idx++
		}
		_ = yaml.Unmarshal([]byte(strings.Join(fm, "\n")), &meta)
	}

	title := "Untitled Survey"
	if t, ok := meta["title"].(string); ok && t != "" {
		title = t
	}
	lang := "EN"
	if l, ok := meta["language"].(string); ok && l != "" {
		lang = strings.ToUpper(l)
	}
	desc := ""
	if d, ok := meta["description"].(string); ok {
		desc = d
	}
	var extraLangs []string
	if rawLangs, ok := meta["languages"]; ok {
		switch v := rawLangs.(type) {
		case string:
			for _, part := range strings.Split(v, ",") {
				p := strings.TrimSpace(part)
				if p != "" {
					extraLangs = append(extraLangs, strings.ToUpper(p))
				}
			}
		case []any:
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					extraLangs = append(extraLangs, strings.ToUpper(s))
				}
			}
		}
	}

	spec := &SurveySpec{
		Title:          title,
		Language:       lang,
		Description:    desc,
		ExtraLanguages: extraLangs,
		Blocks:         []*BlockSpec{},
	}

	var currentBlock *BlockSpec
	var currentQuestion *QuestionSpec

	flushQuestion := func() {
		if currentQuestion != nil {
			if currentBlock == nil {
				currentBlock = &BlockSpec{Name: "Default Question Block", Questions: []*QuestionSpec{}}
			}
			currentBlock.Questions = append(currentBlock.Questions, currentQuestion)
			currentQuestion = nil
		}
	}

	flushBlock := func() {
		flushQuestion()
		if currentBlock != nil {
			spec.Blocks = append(spec.Blocks, currentBlock)
			currentBlock = nil
		}
	}

	ensureBlock := func() {
		if currentBlock == nil {
			currentBlock = &BlockSpec{Name: "Default Question Block", Questions: []*QuestionSpec{}}
		}
	}

	var lastItem trackedItem

	for idx < len(lines) {
		line := lines[idx]
		idx++
		stripped := strings.TrimSpace(line)

		if strings.HasPrefix(stripped, "<!--") {
			continue
		}

		if reBlock.MatchString(line) && !strings.HasPrefix(line, "##") {
			flushBlock()
			currentBlock = &BlockSpec{
				Name:      strings.TrimSpace(strings.TrimPrefix(line, "#")),
				Questions: []*QuestionSpec{},
			}
			continue
		}

		if strings.HasPrefix(strings.ToLower(stripped), "branch-if:") {
			cond := parseLogicCondition(stripped)
			if cond != nil && currentBlock != nil && currentQuestion == nil {
				currentBlock.BranchLogic = cond
			}
			continue
		}

		if strings.HasPrefix(strings.ToLower(stripped), "loop-from:") {
			m := reLoopFrom.FindStringSubmatch(stripped)
			if len(m) >= 2 && currentBlock != nil && currentQuestion == nil {
				raw := m[1]
				if !strings.HasPrefix(raw, "@") {
					raw = strings.ToUpper(raw)
				}
				currentBlock.LoopFrom = raw
			}
			continue
		}

		if strings.HasPrefix(strings.ToLower(stripped), "carry-from:") {
			m := reLoopFrom.FindStringSubmatch(stripped)
			if len(m) >= 2 && currentQuestion != nil {
				raw := m[1]
				if !strings.HasPrefix(raw, "@") {
					raw = strings.ToUpper(raw)
				}
				currentQuestion.CarryFrom = raw
			}
			continue
		}

		if strings.HasPrefix(strings.ToLower(stripped), "show-if:") {
			cond := parseLogicCondition(stripped)
			if cond != nil {
				if currentQuestion != nil {
					currentQuestion.DisplayLogic = cond
				} else if currentBlock != nil {
					currentBlock.DisplayLogic = cond
				}
			}
			continue
		}

		if strings.HasPrefix(strings.ToLower(stripped), "skip-if:") {
			rule := parseSkipRule(stripped)
			if rule != nil && currentQuestion != nil {
				currentQuestion.SkipLogic = append(currentQuestion.SkipLogic, rule)
			}
			continue
		}

		m := reQuestion.FindStringSubmatch(line)
		if len(m) >= 3 {
			flushQuestion()
			ensureBlock()
			required := len(m) > 3 && m[3] == "*"
			currentQuestion = newEmptyQuestion(strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), required)
			if len(m) > 4 && m[4] != "" {
				currentQuestion.Label = m[4]
			}
			lastItem = trackedItem{kind: kindQuestionText}
			continue
		}

		if stripped == "---" {
			flushQuestion()
			ensureBlock()
			currentBlock.Questions = append(currentBlock.Questions, &QuestionSpec{IsPageBreak: true})
			continue
		}

		if currentQuestion == nil {
			continue
		}

		if strings.HasPrefix(strings.ToLower(stripped), "scale:") {
			raw := strings.TrimSpace(stripped[6:])
			parts := strings.Split(raw, ",")
			var scale []string
			for _, p := range parts {
				ps := strings.TrimSpace(p)
				if ps != "" {
					scale = append(scale, ps)
				}
			}
			currentQuestion.Scale = scale
			lastItem = trackedItem{kind: kindScale}
			continue
		}

		mLang := reLangLine.FindStringSubmatch(stripped)
		if len(mLang) >= 4 {
			langCode := strings.ToUpper(mLang[1])
			isScale := mLang[2] != ""
			tText := strings.TrimSpace(mLang[3])
			if isScale {
				var list []string
				for _, p := range strings.Split(tText, ",") {
					ps := strings.TrimSpace(p)
					if ps != "" {
						list = append(list, ps)
					}
				}
				currentQuestion.ScaleTranslations[langCode] = list
			} else {
				switch lastItem.kind {
				case kindQuestionText:
					currentQuestion.TextTranslations[langCode] = tText
				case kindScale:
					var list []string
					for _, p := range strings.Split(tText, ",") {
						ps := strings.TrimSpace(p)
						if ps != "" {
							list = append(list, ps)
						}
					}
					currentQuestion.ScaleTranslations[langCode] = list
				case kindChoice:
					if currentQuestion.ChoiceTranslations[langCode] == nil {
						currentQuestion.ChoiceTranslations[langCode] = make(map[int]string)
					}
					currentQuestion.ChoiceTranslations[langCode][lastItem.index] = tText
				case kindRow:
					if currentQuestion.RowTranslations[langCode] == nil {
						currentQuestion.RowTranslations[langCode] = make(map[int]string)
					}
					currentQuestion.RowTranslations[langCode][lastItem.index] = tText
				}
			}
			continue
		}

		if strings.HasPrefix(stripped, "- ") {
			val := strings.TrimSpace(stripped[2:])
			hasTextEntry := reTextEntry.MatchString(val)
			if hasTextEntry {
				val = strings.TrimSpace(reTextEntry.ReplaceAllString(val, ""))
			}
			var varname string
			var recode string
			ann := reChoiceAnnotation.FindStringSubmatch(val)
			if len(ann) >= 3 {
				val = strings.TrimSpace(ann[1])
				varname = ann[2]
				if len(ann) > 3 {
					recode = ann[3]
				}
			}

			if strings.HasPrefix(currentQuestion.Type, "matrix") {
				currentQuestion.Rows = append(currentQuestion.Rows, val)
				itemIdx := len(currentQuestion.Rows)
				lastItem = trackedItem{kind: kindRow, index: itemIdx}
				if varname != "" {
					currentQuestion.VariableNames[itemIdx] = varname
				}
				if recode != "" {
					currentQuestion.RecodeValues[itemIdx] = recode
				}
			} else {
				currentQuestion.Choices = append(currentQuestion.Choices, val)
				itemIdx := len(currentQuestion.Choices)
				lastItem = trackedItem{kind: kindChoice, index: itemIdx}
				if varname != "" {
					currentQuestion.VariableNames[itemIdx] = varname
				}
				if recode != "" {
					currentQuestion.RecodeValues[itemIdx] = recode
				}
				if hasTextEntry {
					currentQuestion.TextEntryChoices[itemIdx] = true
				}
			}
			continue
		}

		if currentQuestion.Type == "description" {
			currentQuestion.BodyLines = append(currentQuestion.BodyLines, line)
			continue
		}
	}

	flushBlock()
	return spec, nil
}

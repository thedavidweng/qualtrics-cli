package qsf

import (
	"fmt"
	"strings"
)

func shortText(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func validateSpec(spec *SurveySpec) {
	seenLabels := make(map[string]string)
	for _, block := range spec.Blocks {
		for _, q := range block.Questions {
			if q.IsPageBreak {
				continue
			}
			name := shortText(q.Text, 60)
			if _, ok := questionTypes[strings.ToLower(q.Type)]; !ok && q.Type != "likert" {
				msg := fmt.Sprintf("unknown question type [%s] on %q, falling back to [text]", q.Type, name)
				if len(q.Choices) > 0 {
					msg += fmt.Sprintf("; %d choice(s) will be ignored", len(q.Choices))
				}
				spec.Warnings = append(spec.Warnings, msg)
				continue
			}
			switch {
			case q.Type == "mc" || q.Type == "mc-multi" || q.Type == "mc-dropdown" || q.Type == "rank":
				if len(q.Choices) == 0 && q.CarryFrom == "" {
					spec.Warnings = append(spec.Warnings, fmt.Sprintf("question has no choices: %q", name))
				}
			case strings.HasPrefix(q.Type, "matrix") || q.Type == "likert":
				if len(q.Rows) == 0 {
					spec.Warnings = append(spec.Warnings, fmt.Sprintf("matrix question has no rows: %q", name))
				}
				if len(q.Scale) == 0 {
					spec.Warnings = append(spec.Warnings, fmt.Sprintf("matrix question has no scale: %q", name))
				}
				if q.Type == "likert" {
					if len(q.ExclusiveChoices) > 0 {
						spec.Warnings = append(spec.Warnings, fmt.Sprintf("[exclusive] on [likert] %q has no effect; each item becomes its own question", name))
					}
					if len(q.TextEntryChoices) > 0 {
						spec.Warnings = append(spec.Warnings, fmt.Sprintf("[+text] on [likert] rows of %q has no effect; ignoring", name))
					}
					if len(q.RecodeValues) > 0 {
						spec.Warnings = append(spec.Warnings, fmt.Sprintf("recode values on [likert] rows of %q have no effect; use [VARNAME] for export tags", name))
					}
					if q.MinAnswers > 0 || q.MaxAnswers > 0 {
						spec.Warnings = append(spec.Warnings, fmt.Sprintf("min/max-answers on [likert] %q has no effect; each item is single-answer", name))
					}
					if q.CarryFrom != "" {
						spec.Warnings = append(spec.Warnings, fmt.Sprintf("carry-from on [likert] %q has no effect; items come from rows", name))
					}
				}
			case q.Type == "description":
				if strings.TrimSpace(q.Text) == "" && strings.TrimSpace(joinBody(q.BodyLines)) == "" {
					spec.Warnings = append(spec.Warnings, "description question has neither heading nor body text; it will import blank")
				}
			}
			if q.Label != "" {
				if prev, dup := seenLabels[q.Label]; dup {
					spec.Warnings = append(spec.Warnings, fmt.Sprintf("duplicate label @%s on %q (first used on %q); export tags must be unique", q.Label, name, prev))
				} else {
					seenLabels[q.Label] = name
				}
			}
			if q.MinAnswers > 0 || q.MaxAnswers > 0 {
				if q.Type != "mc-multi" && q.Type != "likert" {
					spec.Warnings = append(spec.Warnings, fmt.Sprintf("min/max-answers on %q only applies to [mc-multi]; ignoring", name))
				} else {
					if q.MinAnswers > 0 && q.MaxAnswers > 0 && q.MinAnswers > q.MaxAnswers {
						spec.Warnings = append(spec.Warnings, fmt.Sprintf("min-answers %d exceeds max-answers %d on %q; ignoring both", q.MinAnswers, q.MaxAnswers, name))
					} else {
						if q.MaxAnswers > len(q.Choices) && q.CarryFrom == "" {
							spec.Warnings = append(spec.Warnings, fmt.Sprintf("max-answers %d exceeds %d choices on %q; ignoring", q.MaxAnswers, len(q.Choices), name))
						} else {
							limit := fmt.Sprintf("at most %d", q.MaxAnswers)
							if q.MinAnswers > 0 && q.MaxAnswers > 0 && q.MinAnswers == q.MaxAnswers {
								limit = fmt.Sprintf("exactly %d", q.MaxAnswers)
							} else if q.MinAnswers > 0 && q.MaxAnswers == 0 {
								limit = fmt.Sprintf("at least %d", q.MinAnswers)
							}
							spec.Warnings = append(spec.Warnings, fmt.Sprintf("answer-count limit (%s) on %q has no verified QSF representation and is not enforced; keep the limit in the question text and set Response Requirements manually after import", limit, name))
						}
					}
				}
			}
			if len(q.ExclusiveChoices) > 0 && (q.Type == "text" || q.Type == "text-essay" || q.Type == "description") {
				spec.Warnings = append(spec.Warnings, fmt.Sprintf("[exclusive] on %q has no choices to apply to; ignoring", name))
			}
		}
	}

	likertLabels := make(map[string]int)
	for _, block := range spec.Blocks {
		for _, q := range block.Questions {
			if !q.IsPageBreak && q.Type == "likert" && q.Label != "" {
				likertLabels["@"+q.Label] = len(q.Rows)
			}
		}
	}
	if len(likertLabels) > 0 {
		refHits := func(ref, kind string) {
			if n, ok := likertLabels[ref]; ok {
				spec.Warnings = append(spec.Warnings, fmt.Sprintf("%s references @%s which expands to %d questions; only the first is targeted", kind, ref[1:], n))
			}
		}
		for _, block := range spec.Blocks {
			if block.BranchLogic != nil {
				refHits(block.BranchLogic.QID, "branch-if")
			}
			if block.LoopFrom != "" {
				refHits(block.LoopFrom, "loop-from")
			}
			if block.DisplayLogic != nil {
				refHits(block.DisplayLogic.QID, "show-if")
			}
			for _, q := range block.Questions {
				if q.IsPageBreak {
					continue
				}
				name := shortText(q.Text, 60)
				if q.DisplayLogic != nil {
					refHits(q.DisplayLogic.QID, "show-if on "+name)
				}
				if q.CarryFrom != "" {
					refHits(q.CarryFrom, "carry-from on "+name)
				}
				for _, sl := range q.SkipLogic {
					refHits(sl.Destination, "skip-if on "+name)
				}
			}
		}
	}
}

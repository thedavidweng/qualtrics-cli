package qsf

import (
	"strings"
	"testing"
)

func mustParse(t *testing.T, text string) *SurveySpec {
	t.Helper()
	spec, err := ParseSurvey(text)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func hasWarning(warnings []string, substr string) bool {
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}

func TestDescriptionHeadingFallback(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# Intro\n\n## Long intro heading text here [description]\n\n# Next\n\n## Q? [mc]\n- A\n")
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	elements, ok := qsf["SurveyElements"].([]any)
	if !ok {
		t.Fatal("missing SurveyElements")
	}
	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			t.Fatal("bad element shape")
		}
		if m["Element"] == "SQ" {
			p, ok := m["Payload"].(map[string]any)
			if !ok {
				t.Fatal("bad payload shape")
			}
			if p["QuestionType"] == "DB" {
				found = true
				if p["QuestionText"] != "Long intro heading text here" {
					t.Fatalf("QuestionText = %q, want heading fallback", p["QuestionText"])
				}
			}
		}
	}
	if !found {
		t.Fatal("no DB question in output")
	}
}

func TestDescriptionBodyPreferred(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# Intro\n\n## Label [description]\nBody paragraph one.\n\nBody paragraph two.\n")
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	elements, ok := qsf["SurveyElements"].([]any)
	if !ok {
		t.Fatal("missing SurveyElements")
	}
	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			t.Fatal("bad element shape")
		}
		if m["Element"] == "SQ" {
			p, ok := m["Payload"].(map[string]any)
			if !ok {
				t.Fatal("bad payload shape")
			}
			if p["QuestionType"] == "DB" {
				if p["QuestionText"] != "Body paragraph one.<br><br>Body paragraph two." {
					t.Fatalf("QuestionText = %q", p["QuestionText"])
				}
				return
			}
		}
	}
	t.Fatal("no DB question in output")
}

func TestUnknownTypeWarning(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## What? [slider]\n- A\n- B\n")
	if !hasWarning(spec.Warnings, "unknown question type [slider]") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
	if !hasWarning(spec.Warnings, "will be ignored") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	elements, ok := qsf["SurveyElements"].([]any)
	if !ok {
		t.Fatal("missing SurveyElements")
	}
	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			t.Fatal("bad element shape")
		}
		if m["Element"] == "SQ" {
			p, ok := m["Payload"].(map[string]any)
			if !ok {
				t.Fatal("bad payload shape")
			}
			if p["QuestionType"] != "TE" {
				t.Fatalf("QuestionType = %v, want TE fallback", p["QuestionType"])
			}
		}
	}
}

func TestChoiceValidationWarnings(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Empty mc? [mc]\n\n## Empty matrix? [matrix]\n- Only row\n")
	if !hasWarning(spec.Warnings, "has no choices") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
	if !hasWarning(spec.Warnings, "has no scale") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

func TestMatrixNoRowsWarning(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Rate? [matrix]\nscale: Good, Bad\n")
	if !hasWarning(spec.Warnings, "has no rows") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

func TestDuplicateLabelWarning(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## First? [mc] @same\n- A\n\n## Second? [mc] @same\n- B\n")
	if !hasWarning(spec.Warnings, "duplicate label @same") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

func TestUnresolvedLabelWarning(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## First? [mc] @real\n- A\n- B\n\n# C\nbranch-if: @ghost/1 Selected\n\n## Second? [mc]\n- C\n")
	if _, err := BuildQSF(spec); err != nil {
		t.Fatal(err)
	}
	if !hasWarning(spec.Warnings, "unresolved label @ghost") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

func TestSampleSurveyHasNoWarnings(t *testing.T) {
	spec := mustParse(t, sampleSurvey)
	if _, err := BuildQSF(spec); err != nil {
		t.Fatal(err)
	}
	if len(spec.Warnings) != 0 {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

func sqPayload(t *testing.T, qsf map[string]any, qid string) map[string]any {
	t.Helper()
	elements, ok := qsf["SurveyElements"].([]any)
	if !ok {
		t.Fatal("missing SurveyElements")
	}
	for _, el := range elements {
		m, ok := el.(map[string]any)
		if !ok {
			t.Fatal("bad element shape")
		}
		if m["Element"] == "SQ" && m["PrimaryAttribute"] == qid {
			p, ok := m["Payload"].(map[string]any)
			if !ok {
				t.Fatal("bad payload shape")
			}
			return p
		}
	}
	t.Fatalf("question %s not found", qid)
	return nil
}

func choiceMap(t *testing.T, choices map[string]any, key string) map[string]any {
	t.Helper()
	raw, ok := choices[key]
	if !ok {
		t.Fatalf("choice %s missing", key)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("choice %s has bad shape", key)
	}
	return m
}

func validationSettings(t *testing.T, p map[string]any) map[string]any {
	t.Helper()
	v, ok := p["Validation"].(map[string]any)
	if !ok {
		t.Fatal("missing Validation")
	}
	s, ok := v["Settings"].(map[string]any)
	if !ok {
		t.Fatal("missing Validation.Settings")
	}
	return s
}

func TestValidationMatrix(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Req? [mc]* @r\n- A\n- B\n\n## Opt? [mc] @o\n- A\n\n## Essay? [text-essay]\n\n## Intro [description]\nBody.\n")
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	s := validationSettings(t, sqPayload(t, qsf, "QID1"))
	if s["ForceResponse"] != "ON" || s["ForceResponseType"] != "ON" || s["Type"] != "None" {
		t.Fatalf("required validation = %v", s)
	}
	s = validationSettings(t, sqPayload(t, qsf, "QID2"))
	if s["ForceResponse"] != "OFF" || s["ForceResponseType"] != "ON" || s["Type"] != "None" {
		t.Fatalf("optional validation = %v", s)
	}
	s = validationSettings(t, sqPayload(t, qsf, "QID4"))
	if len(s) != 1 || s["Type"] != "None" {
		t.Fatalf("description validation = %v", s)
	}
}

func TestExclusiveChoice(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Pick? [mc-multi]\n- A\n- None [exclusive]\n- Other [+text] [exclusive]\n")
	if len(spec.Warnings) != 0 {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	p := sqPayload(t, qsf, "QID1")
	choices, ok := p["Choices"].(map[string]any)
	if !ok {
		t.Fatal("missing Choices")
	}
	if _, present := choiceMap(t, choices, "1")["ExclusiveAnswer"]; present {
		t.Fatal("unmarked choice must not carry ExclusiveAnswer")
	}
	if choiceMap(t, choices, "2")["ExclusiveAnswer"] != true {
		t.Fatalf("choice 2 = %v", choices["2"])
	}
	c3 := choiceMap(t, choices, "3")
	if c3["ExclusiveAnswer"] != true || c3["TextEntry"] != "true" || c3["Display"] != "Other" {
		t.Fatalf("choice 3 = %v", c3)
	}
}

func TestMatrixExclusiveRows(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Rate? [matrix]\nscale: Good, Bad\n- Speed\n- None [exclusive]\n")
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	p := sqPayload(t, qsf, "QID1")
	choices, ok := p["Choices"].(map[string]any)
	if !ok {
		t.Fatal("missing Choices")
	}
	if choiceMap(t, choices, "1")["ExclusiveAnswer"] != false {
		t.Fatalf("row 1 = %v", choices["1"])
	}
	if choiceMap(t, choices, "2")["ExclusiveAnswer"] != true {
		t.Fatalf("row 2 = %v", choices["2"])
	}
	answers, ok := p["Answers"].(map[string]any)
	if !ok {
		t.Fatal("missing Answers")
	}
	if _, present := choiceMap(t, answers, "1")["ExclusiveAnswer"]; present {
		t.Fatal("scale answers must not carry ExclusiveAnswer")
	}
}

func TestMinMaxAnswers(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Pick up to two? [mc-multi]\nmax-answers: 2\n- A\n- B\n- C\n")
	if !hasWarning(spec.Warnings, "at most 2") || !hasWarning(spec.Warnings, "manually after import") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
	if spec.Blocks[0].Questions[0].MaxAnswers != 2 {
		t.Fatal("MaxAnswers not parsed")
	}

	spec = mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Pick? [mc-multi]\nmin-answers: 3\nmax-answers: 2\n- A\n- B\n")
	if !hasWarning(spec.Warnings, "min-answers 3 exceeds max-answers 2") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}

	spec = mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Pick? [mc-multi]\nmax-answers: 9\n- A\n- B\n")
	if !hasWarning(spec.Warnings, "exceeds 2 choices") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}

	spec = mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Pick? [mc]\nmax-answers: 1\n- A\n- B\n")
	if !hasWarning(spec.Warnings, "only applies to [mc-multi]") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

func TestLikertExpansion(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## How important is this? [likert]* @imp\nscale: Low, High\n- Speed [SPEED]\n- Quality\n")
	if len(spec.Warnings) != 0 {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}
	p1 := sqPayload(t, qsf, "QID1")
	if p1["QuestionType"] != "MC" || p1["Selector"] != "SAVR" {
		t.Fatalf("QID1 = %v %v", p1["QuestionType"], p1["Selector"])
	}
	if p1["QuestionText"] != "How important is this? Speed" {
		t.Fatalf("QID1 text = %q", p1["QuestionText"])
	}
	if p1["DataExportTag"] != "SPEED" {
		t.Fatalf("QID1 tag = %v", p1["DataExportTag"])
	}
	choices, ok := p1["Choices"].(map[string]any)
	if !ok || len(choices) != 2 {
		t.Fatalf("QID1 choices = %v", p1["Choices"])
	}
	p2 := sqPayload(t, qsf, "QID2")
	if p2["QuestionText"] != "How important is this? Quality" {
		t.Fatalf("QID2 text = %q", p2["QuestionText"])
	}
	if p2["DataExportTag"] != "imp_2" {
		t.Fatalf("QID2 tag = %v", p2["DataExportTag"])
	}
	s := validationSettings(t, p1)
	if s["ForceResponse"] != "ON" || s["ForceResponseType"] != "ON" {
		t.Fatalf("QID1 validation = %v", s)
	}
}

func TestLikertValidation(t *testing.T) {
	spec := mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## Rate? [likert]\n- Only row\n")
	if !hasWarning(spec.Warnings, "has no scale") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}

	spec = mustParse(t, "---\ntitle: T\n---\n\n# B\n\n## First? [likert] @a\nscale: Good, Bad\n- X\n- Y\n\n## Second? [mc]\nshow-if: @a/1 Selected\n- Z\n")
	if _, err := BuildQSF(spec); err != nil {
		t.Fatal(err)
	}
	if !hasWarning(spec.Warnings, "expands to 2 questions") {
		t.Fatalf("warnings = %v", spec.Warnings)
	}
}

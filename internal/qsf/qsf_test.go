package qsf

import (
	"encoding/json"
	"strings"
	"testing"
)

const sampleSurvey = `---
title: Sample Research Survey
language: EN
description: A test survey for research.
languages: ES, FR
---

# Demographics

## What is your age group? [mc]* @age
- Under 18
- 18-24 [VAR_AGE_YOUNG=1]
- 25-34 [VAR_AGE_MID=2]
- 35 or older
skip-if: 1 Selected → ENDOFSURVEY

## Please describe your background [text-essay]
lang-es: Describa sus antecedentes

---

## What tools do you use? [mc-multi] @tools
- Python
- Go
- Other [+text]

# Experience
branch-if: @age/2 Selected

## Rate your experience with Go [matrix]*
scale: Poor, Fair, Good, Excellent
lang-es-scale: Malo, Regular, Bueno, Excelente
- Compiler speed
- Standard library
- Tooling

# Follow Up
loop-from: @tools

## How often do you use ${lm://Field/1}? [mc]
- Daily
- Weekly
- Rarely
`

func TestParseSurvey(t *testing.T) {
	spec, err := ParseSurvey(sampleSurvey)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Title != "Sample Research Survey" {
		t.Fatalf("title = %q", spec.Title)
	}
	if spec.Language != "EN" {
		t.Fatalf("lang = %q", spec.Language)
	}
	if len(spec.ExtraLanguages) != 2 || spec.ExtraLanguages[0] != "ES" || spec.ExtraLanguages[1] != "FR" {
		t.Fatalf("extra langs = %+v", spec.ExtraLanguages)
	}
	if len(spec.Blocks) != 3 {
		t.Fatalf("blocks = %d, want 3", len(spec.Blocks))
	}

	// Block 1
	b1 := spec.Blocks[0]
	if b1.Name != "Demographics" {
		t.Fatalf("b1 name = %s", b1.Name)
	}
	// Q1: mc, required, label=age, skip-if
	q1 := b1.Questions[0]
	if q1.Type != "mc" || !q1.Required || q1.Label != "age" {
		t.Fatalf("q1 = %+v", q1)
	}
	if len(q1.Choices) != 4 || q1.VariableNames[2] != "VAR_AGE_YOUNG" || q1.RecodeValues[2] != "1" {
		t.Fatalf("q1 choices = %+v, recode = %+v", q1.Choices, q1.RecodeValues)
	}
	if len(q1.SkipLogic) != 1 || q1.SkipLogic[0].Destination != "ENDOFSURVEY" {
		t.Fatalf("q1 skip = %+v", q1.SkipLogic)
	}

	// Page break
	if !b1.Questions[2].IsPageBreak {
		t.Fatalf("expected page break at q idx 2")
	}

	// Q3: mc-multi with inline text entry
	q3 := b1.Questions[3]
	if q3.Type != "mc-multi" || !q3.TextEntryChoices[3] {
		t.Fatalf("q3 text entry: %+v", q3.TextEntryChoices)
	}

	// Block 2: branch-if @age/2 Selected
	b2 := spec.Blocks[1]
	if b2.BranchLogic == nil || b2.BranchLogic.QID != "@age" || b2.BranchLogic.Choice != 2 {
		t.Fatalf("b2 branch: %+v", b2.BranchLogic)
	}

	// Block 3: loop-from @tools
	b3 := spec.Blocks[2]
	if b3.LoopFrom != "@tools" {
		t.Fatalf("b3 loopFrom = %s", b3.LoopFrom)
	}
}

func TestBuildQSF(t *testing.T) {
	spec, err := ParseSurvey(sampleSurvey)
	if err != nil {
		t.Fatal(err)
	}

	qsf, err := BuildQSF(spec)
	if err != nil {
		t.Fatal(err)
	}

	entry, ok := qsf["SurveyEntry"].(map[string]any)
	if !ok {
		t.Fatal("missing SurveyEntry")
	}
	if entry["SurveyName"] != "Sample Research Survey" {
		t.Fatalf("survey name = %v", entry["SurveyName"])
	}
	sid, ok := entry["SurveyID"].(string)
	if !ok || !strings.HasPrefix(sid, "SV_") {
		t.Fatalf("survey id = %v", entry["SurveyID"])
	}

	elements, ok := qsf["SurveyElements"].([]any)
	if !ok || len(elements) < 10 {
		t.Fatalf("elements count = %d", len(elements))
	}

	// Verify JSON marshaling
	data, err := json.MarshalIndent(qsf, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	// Summarize
	summary, err := SummarizeQSF(data)
	if err != nil {
		t.Fatal(err)
	}
	if summary.SurveyName != "Sample Research Survey" {
		t.Fatalf("summary name = %s", summary.SurveyName)
	}
	if summary.BlockCount != 3 {
		t.Fatalf("summary blocks = %d", summary.BlockCount)
	}
	if summary.TotalQCount != 5 {
		t.Fatalf("summary questions = %d, want 5", summary.TotalQCount)
	}

	// Convert to definition
	defData, err := ConvertQSFToDefinition(data)
	if err != nil {
		t.Fatal(err)
	}
	var def DefinitionOutput
	if err := json.Unmarshal(defData, &def); err != nil {
		t.Fatal(err)
	}
	if def.SurveyName != "Sample Research Survey" {
		t.Fatalf("def name = %s", def.SurveyName)
	}
	if len(def.Blocks) < 3 {
		t.Fatalf("def blocks = %d", len(def.Blocks))
	}
	if len(def.Questions) != 5 {
		t.Fatalf("def questions = %d, want 5", len(def.Questions))
	}
}

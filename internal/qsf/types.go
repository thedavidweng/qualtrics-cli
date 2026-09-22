package qsf

type SurveySpec struct {
	Title          string
	Language       string
	Description    string
	ExtraLanguages []string
	Blocks         []*BlockSpec
	Warnings       []string
}

type BlockSpec struct {
	Name         string
	BranchLogic  *LogicCondition
	DisplayLogic *LogicCondition
	LoopFrom     string
	Questions    []*QuestionSpec
}

type QuestionSpec struct {
	Text               string
	Type               string
	Required           bool
	Label              string
	Choices            []string
	Rows               []string
	Scale              []string
	BodyLines          []string
	SkipLogic          []*SkipRule
	DisplayLogic       *LogicCondition
	LikertBase         string
	RecodeValues       map[int]string
	VariableNames      map[int]string
	TextEntryChoices   map[int]bool
	ExclusiveChoices   map[int]bool
	MinAnswers         int
	MaxAnswers         int
	CarryFrom          string
	TextTranslations   map[string]string
	ChoiceTranslations map[string]map[int]string
	RowTranslations    map[string]map[int]string
	ScaleTranslations  map[string][]string
	IsPageBreak        bool
}

type LogicCondition struct {
	QID      string `json:"qid"`
	Choice   int    `json:"choice"`
	Operator string `json:"operator"`
}

type SkipRule struct {
	Choice      int    `json:"choice"`
	Condition   string `json:"condition"`
	Destination string `json:"destination"`
}

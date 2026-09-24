package qualtrics

import (
	"context"
	"net/http"
	"testing"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func TestGetDefinitionParsesBlocks(t *testing.T) {
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		return testutil.JSONResponse(200, fixture(t, "definition.json")), nil
	})

	def, err := client.GetDefinition(context.Background(), "SV_1")
	if err != nil {
		t.Fatal(err)
	}
	summary := def.Summary()
	if summary["blocks"] != 2 {
		t.Fatalf("blocks = %v", summary["blocks"])
	}
	if summary["questions"] != 3 {
		t.Fatalf("questions = %v", summary["questions"])
	}
	perBlock, _ := summary["per_block"].([]map[string]any)
	if len(perBlock) != 2 {
		t.Fatalf("per_block = %+v", perBlock)
	}
}

func TestListQuestionsAcceptsArrayAndMap(t *testing.T) {
	bodies := []string{
		fixture(t, "questions_array.json"),
		fixture(t, "questions_map.json"),
	}
	for _, body := range bodies {
		client := stub(t, func(r *http.Request) (*http.Response, error) {
			return testutil.JSONResponse(200, body), nil
		})
		questions, err := client.ListQuestions(context.Background(), "SV_1")
		if err != nil {
			t.Fatal(err)
		}
		if len(questions) != 1 || questions[0].QuestionID != "QID1" {
			t.Fatalf("questions = %+v", questions)
		}
	}
}

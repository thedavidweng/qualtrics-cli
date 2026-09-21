package qualtrics

import (
	"context"
	"io"
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

func TestCreateQuestionRequestShape(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotBody string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return testutil.JSONResponse(200, `{"result":{"QuestionID":"QID9"}}`), nil
	})

	q, err := client.CreateQuestion(context.Background(), "SV_1", "BL_2", []byte(`{"QuestionText":"Hi","QuestionType":"MC"}`))
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" || gotPath != "/API/v3/survey-definitions/SV_1/questions" {
		t.Fatalf("%s %s", gotMethod, gotPath)
	}
	if gotQuery != "blockId=BL_2" {
		t.Fatalf("query = %s", gotQuery)
	}
	if gotBody != `{"QuestionText":"Hi","QuestionType":"MC"}` {
		t.Fatalf("body = %s", gotBody)
	}
	if q.QuestionID != "QID9" {
		t.Fatalf("qid = %s", q.QuestionID)
	}
}

func TestCreateBlockRequestShape(t *testing.T) {
	var gotBody string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return testutil.JSONResponse(200, `{"result":{"ID":"BL_9","Type":"Standard"}}`), nil
	})

	b, err := client.CreateBlock(context.Background(), "SV_1", "Demographics")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"Description":"Demographics","Type":"Standard"}`
	if gotBody != want {
		t.Fatalf("body = %s, want %s", gotBody, want)
	}
	if b.ID != "BL_9" {
		t.Fatalf("id = %s", b.ID)
	}
}

func TestQuestionBlockFlowOptionShapes(t *testing.T) {
	type call struct {
		method string
		path   string
	}
	var got call
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		got = call{r.Method, r.URL.Path}
		return testutil.JSONResponse(200, `{"result":{"ok":true}}`), nil
	})
	ctx := context.Background()

	if err := client.UpdateQuestion(ctx, "SV_1", "QID1", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if got != (call{"PUT", "/API/v3/survey-definitions/SV_1/questions/QID1"}) {
		t.Fatalf("update question: %+v", got)
	}

	if err := client.DeleteQuestion(ctx, "SV_1", "QID1"); err != nil {
		t.Fatal(err)
	}
	if got != (call{"DELETE", "/API/v3/survey-definitions/SV_1/questions/QID1"}) {
		t.Fatalf("delete question: %+v", got)
	}

	if err := client.UpdateBlock(ctx, "SV_1", "BL_1", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if got != (call{"PUT", "/API/v3/survey-definitions/SV_1/blocks/BL_1"}) {
		t.Fatalf("update block: %+v", got)
	}

	if err := client.UpdateFlow(ctx, "SV_1", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if got != (call{"PUT", "/API/v3/survey-definitions/SV_1/flow"}) {
		t.Fatalf("update flow: %+v", got)
	}

	if err := client.UpdateOptions(ctx, "SV_1", []byte(`{}`)); err != nil {
		t.Fatal(err)
	}
	if got != (call{"PUT", "/API/v3/survey-definitions/SV_1/options"}) {
		t.Fatalf("update options: %+v", got)
	}
}

func TestCreateFromDefinitionPostsPayload(t *testing.T) {
	var gotBody string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return testutil.JSONResponse(200, `{"result":{"SurveyID":"SV_imp"}}`), nil
	})

	payload := []byte(`{"SurveyName":"Imported"}`)
	def, err := client.CreateFromDefinition(context.Background(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if gotBody != string(payload) {
		t.Fatalf("body = %s", gotBody)
	}
	if def.SurveyID != "SV_imp" {
		t.Fatalf("id = %s", def.SurveyID)
	}
}

package qualtrics

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func stub(t *testing.T, fn func(*http.Request) (*http.Response, error)) *Client {
	t.Helper()
	c := NewClient("https://pdx1.qualtrics.com/API/v3", "tok", 5*time.Second)
	c.HTTP.Transport = testutil.RoundTripFunc(fn)
	return c
}

func TestListSurveysRequestShape(t *testing.T) {
	var gotPath, gotQuery string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		return testutil.JSONResponse(200, `{"result":{"elements":[{"SurveyID":"SV_1","SurveyName":"N"}]},"meta":{"httpStatus":"200 - OK"}}`), nil
	})

	page, err := client.ListSurveys(context.Background(), ListOptions{Offset: 50, Limit: 25})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/API/v3/surveys" {
		t.Fatalf("path = %s", gotPath)
	}
	if gotQuery != "limit=25&offset=50" {
		t.Fatalf("query = %s", gotQuery)
	}
	if len(page.Items) != 1 || page.Items[0].SurveyID != "SV_1" {
		t.Fatalf("items = %+v", page.Items)
	}
	if !page.HasMore || page.NextOffset != 51 {
		t.Fatalf("page = %+v", page)
	}
}

func TestGetSurveyRequestShape(t *testing.T) {
	var gotMethod, gotPath string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		return testutil.JSONResponse(200, `{"result":{"SurveyID":"SV_abc123","SurveyName":"N","SurveyStatus":"Active"}}`), nil
	})

	s, err := client.GetSurvey(context.Background(), "SV_abc123")
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "GET" || gotPath != "/API/v3/surveys/SV_abc123" {
		t.Fatalf("%s %s", gotMethod, gotPath)
	}
	if s.SurveyStatus != "Active" {
		t.Fatalf("status = %s", s.SurveyStatus)
	}
}

func TestCreateSurveyRequestShape(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return testutil.JSONResponse(200, `{"result":{"SurveyID":"SV_new","SurveyName":"Fresh"}}`), nil
	})

	s, err := client.CreateSurvey(context.Background(), CreateSurveyRequest{
		SurveyName:      "Fresh",
		Language:        "EN",
		ProjectCategory: "CORE",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != "POST" || gotPath != "/API/v3/survey-definitions" {
		t.Fatalf("%s %s", gotMethod, gotPath)
	}
	want := `{"SurveyName":"Fresh","Language":"EN","ProjectCategory":"CORE"}`
	if gotBody != want {
		t.Fatalf("body = %s, want %s", gotBody, want)
	}
	if s.SurveyID != "SV_new" {
		t.Fatalf("id = %s", s.SurveyID)
	}
}

func TestDeleteSurveyRequestShape(t *testing.T) {
	var gotMethod, gotPath string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		return testutil.JSONResponse(200, `{"result":null}`), nil
	})

	if err := client.DeleteSurvey(context.Background(), "SV_abc123"); err != nil {
		t.Fatal(err)
	}
	if gotMethod != "DELETE" || gotPath != "/API/v3/survey-definitions/SV_abc123" {
		t.Fatalf("%s %s", gotMethod, gotPath)
	}
}

func TestCollectAllWalksPages(t *testing.T) {
	calls := 0
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		calls++
		switch r.URL.Query().Get("offset") {
		case "", "0":
			return testutil.JSONResponse(200, `{"result":{"elements":[{"SurveyID":"SV_1"},{"SurveyID":"SV_2"}]}}`), nil
		default:
			return testutil.JSONResponse(200, `{"result":{"elements":[]}}`), nil
		}
	})

	all, err := CollectAll(context.Background(), client.ListSurveys, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("collected %d", len(all))
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

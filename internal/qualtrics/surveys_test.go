package qualtrics

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func stub(t *testing.T, fn func(*http.Request) (*http.Response, error)) *Client {
	t.Helper()
	c := NewClient("https://pdx1.qualtrics.com/API/v3", "tok", 5*time.Second)
	c.HTTP.Transport = testutil.RoundTripFunc(fn)
	return c
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

func TestPaginateWalksPages(t *testing.T) {
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

	all, err := Paginate(context.Background(), client.ListSurveys, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || calls != 2 {
		t.Fatalf("collected %d in %d calls", len(all), calls)
	}
}

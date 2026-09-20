package qualtrics

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	qerrors "github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func newTestClient(fn testutil.RoundTripFunc) *Client {
	c := NewClient("https://pdx1.qualtrics.com/API/v3", "tok", 5*time.Second)
	c.HTTP.Transport = fn
	return c
}

func TestDoUnwrapsResult(t *testing.T) {
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("X-API-TOKEN") != "tok" {
			t.Errorf("missing token header")
		}
		if !strings.HasSuffix(r.URL.Path, "/API/v3/surveys") {
			t.Errorf("path = %s", r.URL.Path)
		}
		return testutil.JSONResponse(200, `{"result":{"elements":[{"SurveyID":"SV_1"}]},"meta":{"httpStatus":"200 - OK"}}`), nil
	})

	var out struct {
		Elements []struct {
			SurveyID string `json:"SurveyID"`
		} `json:"elements"`
	}
	if err := client.Do(context.Background(), "GET", "/surveys", nil, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Elements) != 1 || out.Elements[0].SurveyID != "SV_1" {
		t.Fatalf("unexpected payload: %+v", out)
	}
}

func TestDoRejectsRedirects(t *testing.T) {
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 302,
			Header:     http.Header{"Location": []string{"https://evil.example.com"}},
			Body:       http.NoBody,
		}, nil
	})
	var out any
	err := client.Do(context.Background(), "GET", "/surveys", nil, &out)
	if err == nil {
		t.Fatal("expected redirect rejection")
	}
	if !strings.Contains(err.Error(), "302") {
		t.Fatalf("err = %v", err)
	}
}

func TestRetriesOn429WithRetryAfter(t *testing.T) {
	var calls int32
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return testutil.JSONResponse(429, `{"meta":{"httpStatus":"429 - Too Many Requests","error":{"errorCode":"RATE_1","errorMessage":"slow down","retryAfter":"0"}}}`), nil
		}
		return testutil.JSONResponse(200, `{"result":{"ok":true}}`), nil
	})

	var out map[string]any
	if err := client.Do(context.Background(), "GET", "/surveys", nil, &out); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if out["ok"] != true {
		t.Fatalf("payload = %+v", out)
	}
}

func TestNoRetryOnPost(t *testing.T) {
	var calls int32
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return testutil.JSONResponse(500, `{"meta":{"httpStatus":"500 - Internal Server Error","error":{"errorCode":"X_1","errorMessage":"boom"}}}`), nil
	})

	var out any
	err := client.Do(context.Background(), "POST", "/responseexports", map[string]any{"surveyId": "SV_1"}, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("POST should not retry, calls = %d", calls)
	}
}

func TestMapsAuthErrors(t *testing.T) {
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		return testutil.JSONResponse(401, `{"meta":{"httpStatus":"401 - Unauthorized","error":{"errorCode":"DCD_7","errorMessage":"Unrecognized X-API-TOKEN."}}}`), nil
	})
	var out any
	err := client.Do(context.Background(), "GET", "/surveys", nil, &out)
	var e *qerrors.Error
	if !errors.As(err, &e) {
		t.Fatalf("err = %v", err)
	}
	if e.Code != "AUTH_TOKEN_INVALID" {
		t.Fatalf("code = %s", e.Code)
	}
	if e.ExitCode() != 3 {
		t.Fatalf("exit = %d, want 3", e.ExitCode())
	}
}

func TestMapsAuthZ(t *testing.T) {
	client := newTestClient(func(r *http.Request) (*http.Response, error) {
		return testutil.JSONResponse(403, `{"meta":{"httpStatus":"403 - Forbidden","error":{"errorCode":"AuthZ_2.0","errorMessage":"This user does not have access to this API endpoint."}}}`), nil
	})
	var out any
	err := client.Do(context.Background(), "GET", "/surveys", nil, &out)
	var e *qerrors.Error
	if !errors.As(err, &e) {
		t.Fatalf("err = %v", err)
	}
	if e.Code != "API_ACCESS_FORBIDDEN" {
		t.Fatalf("code = %s", e.Code)
	}
}

func TestMapsValidationAndNotFound(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{400, `{"meta":{"httpStatus":"400 - Bad Request","error":{"errorCode":"QVAL_7","errorMessage":"bad surveyId"}}}`, "VALIDATION_FAILED"},
		{404, `{"meta":{"httpStatus":"404 - Not Found","error":{"errorMessage":"not found"}}}`, "RESOURCE_NOT_FOUND"},
	}
	for _, c := range cases {
		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			return testutil.JSONResponse(c.status, c.body), nil
		})
		var out any
		err := client.Do(context.Background(), "GET", "/surveys", nil, &out)
		var e *qerrors.Error
		if !errors.As(err, &e) {
			t.Fatalf("err = %v", err)
		}
		if string(e.Code) != c.want {
			t.Fatalf("code = %s, want %s", e.Code, c.want)
		}
	}
}

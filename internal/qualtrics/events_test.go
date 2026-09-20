package qualtrics

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func TestEventSubscriptionsRequestShapes(t *testing.T) {
	type call struct {
		method string
		path   string
	}
	var got call
	var gotBody string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		got = call{r.Method, r.URL.Path}
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
		}
		return testutil.JSONResponse(200, `{"result":{"id":"SUB_1","topics":"surveyengine.completedResponse.SV_1","publicationUrl":"https://example.com/webhook"}}`), nil
	})
	ctx := context.Background()

	// Create
	sub, err := client.CreateEventSubscription(ctx, []byte(`{"topics":"surveyengine.completedResponse.SV_1","publicationUrl":"https://example.com/webhook"}`))
	if err != nil || got != (call{"POST", "/API/v3/eventsubscriptions"}) || sub.ID != "SUB_1" {
		t.Fatalf("create event sub: %+v, body: %s", got, gotBody)
	}

	// Get
	sub, err = client.GetEventSubscription(ctx, "SUB_1")
	if err != nil || got != (call{"GET", "/API/v3/eventsubscriptions/SUB_1"}) || sub.ID != "SUB_1" {
		t.Fatalf("get event sub: %+v", got)
	}

	// Delete
	err = client.DeleteEventSubscription(ctx, "SUB_1")
	if err != nil || got != (call{"DELETE", "/API/v3/eventsubscriptions/SUB_1"}) {
		t.Fatalf("delete event sub: %+v", got)
	}
}

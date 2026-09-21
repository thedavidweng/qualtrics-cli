package qualtrics

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func TestDistributionsRequestShapes(t *testing.T) {
	type call struct {
		method string
		path   string
		query  string
	}
	var got call
	var gotBody string
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		got = call{r.Method, r.URL.Path, r.URL.RawQuery}
		if r.Body != nil {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
		}
		return testutil.JSONResponse(200, `{"result":{"elements":[{"id":"EMD_1","linkId":"LNK_1"}],"id":"EMD_1","linkId":"LNK_1"}}`), nil
	})
	ctx := context.Background()

	// List distributions
	page, err := client.ListDistributions(ctx, "SV_test", ListOptions{Offset: 10, Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "GET" || got.path != "/API/v3/distributions" || got.query != "limit=5&offset=10&surveyId=SV_test" {
		t.Fatalf("list distributions: %+v", got)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "EMD_1" {
		t.Fatalf("page: %+v", page)
	}

	// Get distribution
	dist, err := client.GetDistribution(ctx, "EMD_1", "SV_test")
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "GET" || got.path != "/API/v3/distributions/EMD_1" || got.query != "surveyId=SV_test" {
		t.Fatalf("get distribution: %+v", got)
	}
	if dist.ID != "EMD_1" {
		t.Fatalf("dist id: %s", dist.ID)
	}

	// Create distribution
	_, err = client.CreateDistribution(ctx, []byte(`{"surveyId":"SV_1","action":"CreateDistribution"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "POST" || got.path != "/API/v3/distributions" || gotBody != `{"surveyId":"SV_1","action":"CreateDistribution"}` {
		t.Fatalf("create distribution: %+v, body: %s", got, gotBody)
	}

	// Delete distribution
	err = client.DeleteDistribution(ctx, "EMD_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "DELETE" || got.path != "/API/v3/distributions/EMD_1" {
		t.Fatalf("delete distribution: %+v", got)
	}

	// List links
	linkPage, err := client.ListDistributionLinks(ctx, "EMD_1", "SV_test", ListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "GET" || got.path != "/API/v3/distributions/EMD_1/links" || got.query != "limit=10&surveyId=SV_test" {
		t.Fatalf("list links: %+v", got)
	}
	if len(linkPage.Items) != 1 || linkPage.Items[0].LinkID != "LNK_1" {
		t.Fatalf("links: %+v", linkPage)
	}

	// Get link
	link, err := client.GetDistributionLink(ctx, "EMD_1", "LNK_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "GET" || got.path != "/API/v3/distributions/EMD_1/links/LNK_1" {
		t.Fatalf("get link: %+v", got)
	}
	if link.LinkID != "LNK_1" {
		t.Fatalf("link id: %s", link.LinkID)
	}

	// Create links
	_, err = client.CreateDistributionLinks(ctx, "EMD_1", []byte(`{"surveyId":"SV_1"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "POST" || got.path != "/API/v3/distributions/EMD_1/links" {
		t.Fatalf("create links: %+v", got)
	}

	// Update link
	err = client.UpdateDistributionLink(ctx, "EMD_1", "LNK_1", []byte(`{"status":"Success"}`))
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "PUT" || got.path != "/API/v3/distributions/EMD_1/links/LNK_1" {
		t.Fatalf("update link: %+v", got)
	}

	// Delete link
	err = client.DeleteDistributionLink(ctx, "EMD_1", "LNK_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.method != "DELETE" || got.path != "/API/v3/distributions/EMD_1/links/LNK_1" {
		t.Fatalf("delete link: %+v", got)
	}
}

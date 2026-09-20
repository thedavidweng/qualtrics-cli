package qualtrics

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func TestDirectoriesRequestShapes(t *testing.T) {
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
		return testutil.JSONResponse(200, `{"result":{"elements":[{"directoryId":"POOL_1","id":"CG_1","contactId":"MLR_1"}],"directoryId":"POOL_1","id":"CG_1","contactId":"MLR_1"}}`), nil
	})
	ctx := context.Background()

	// Directories CRUD
	_, err := client.ListDirectories(ctx, ListOptions{Limit: 10})
	if err != nil || got != (call{"GET", "/API/v3/directories"}) {
		t.Fatalf("list directories: %+v", got)
	}

	d, err := client.GetDirectory(ctx, "POOL_1")
	if err != nil || got != (call{"GET", "/API/v3/directories/POOL_1"}) || d.DirectoryID != "POOL_1" {
		t.Fatalf("get directory: %+v", got)
	}

	_, err = client.CreateDirectory(ctx, []byte(`{"name":"Main"}`))
	if err != nil || got != (call{"POST", "/API/v3/directories"}) || gotBody != `{"name":"Main"}` {
		t.Fatalf("create directory: %+v, body: %s", got, gotBody)
	}

	err = client.UpdateDirectory(ctx, "POOL_1", []byte(`{"name":"Renamed"}`))
	if err != nil || got != (call{"PUT", "/API/v3/directories/POOL_1"}) {
		t.Fatalf("update directory: %+v", got)
	}

	err = client.DeleteDirectory(ctx, "POOL_1")
	if err != nil || got != (call{"DELETE", "/API/v3/directories/POOL_1"}) {
		t.Fatalf("delete directory: %+v", got)
	}

	// MailingLists CRUD
	_, err = client.ListMailingLists(ctx, "POOL_1", ListOptions{})
	if err != nil || got != (call{"GET", "/API/v3/directories/POOL_1/mailinglists"}) {
		t.Fatalf("list mailinglists: %+v", got)
	}

	ml, err := client.GetMailingList(ctx, "POOL_1", "CG_1")
	if err != nil || got != (call{"GET", "/API/v3/directories/POOL_1/mailinglists/CG_1"}) || ml.ID != "CG_1" {
		t.Fatalf("get mailinglist: %+v", got)
	}

	_, err = client.CreateMailingList(ctx, "POOL_1", []byte(`{"name":"List 1"}`))
	if err != nil || got != (call{"POST", "/API/v3/directories/POOL_1/mailinglists"}) {
		t.Fatalf("create mailinglist: %+v", got)
	}

	err = client.UpdateMailingList(ctx, "POOL_1", "CG_1", []byte(`{"name":"List 2"}`))
	if err != nil || got != (call{"PUT", "/API/v3/directories/POOL_1/mailinglists/CG_1"}) {
		t.Fatalf("update mailinglist: %+v", got)
	}

	err = client.DeleteMailingList(ctx, "POOL_1", "CG_1")
	if err != nil || got != (call{"DELETE", "/API/v3/directories/POOL_1/mailinglists/CG_1"}) {
		t.Fatalf("delete mailinglist: %+v", got)
	}

	// Contacts CRUD
	_, err = client.ListContacts(ctx, "POOL_1", "CG_1", ListOptions{})
	if err != nil || got != (call{"GET", "/API/v3/directories/POOL_1/mailinglists/CG_1/contacts"}) {
		t.Fatalf("list contacts: %+v", got)
	}

	c, err := client.GetContact(ctx, "POOL_1", "CG_1", "MLR_1")
	if err != nil || got != (call{"GET", "/API/v3/directories/POOL_1/mailinglists/CG_1/contacts/MLR_1"}) || c.ContactID != "MLR_1" {
		t.Fatalf("get contact: %+v", got)
	}

	_, err = client.CreateContact(ctx, "POOL_1", "CG_1", []byte(`{"email":"test@example.com"}`))
	if err != nil || got != (call{"POST", "/API/v3/directories/POOL_1/mailinglists/CG_1/contacts"}) {
		t.Fatalf("create contact: %+v", got)
	}

	err = client.UpdateContact(ctx, "POOL_1", "CG_1", "MLR_1", []byte(`{"email":"updated@example.com"}`))
	if err != nil || got != (call{"PUT", "/API/v3/directories/POOL_1/mailinglists/CG_1/contacts/MLR_1"}) {
		t.Fatalf("update contact: %+v", got)
	}

	err = client.DeleteContact(ctx, "POOL_1", "CG_1", "MLR_1")
	if err != nil || got != (call{"DELETE", "/API/v3/directories/POOL_1/mailinglists/CG_1/contacts/MLR_1"}) {
		t.Fatalf("delete contact: %+v", got)
	}
}

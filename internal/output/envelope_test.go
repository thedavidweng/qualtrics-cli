package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
)

func TestEnvelopeShape(t *testing.T) {
	env := NewEnvelope("surveys.list", "default", SchemaVersion, "req-1", map[string]any{"id": "SV_1"}, 1500*time.Millisecond)
	env.Meta.Pagination = &PaginationMeta{Limit: 100, Offset: 0, Total: 3, HasMore: false}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)

	for _, want := range []string{
		`"ok":true`,
		`"data":{"id":"SV_1"}`,
		`"command":"surveys.list"`,
		`"profile":"default"`,
		`"duration_ms":1500`,
		`"schema_version":"` + SchemaVersion + `"`,
		`"request_id":"req-1"`,
		`"pagination":{"limit":100,"total":3}`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("envelope missing %s: %s", want, got)
		}
	}
}

func TestErrorEnvelopeShape(t *testing.T) {
	env := NewErrorEnvelope("surveys.list", "default", SchemaVersion, errors.New(errors.APIAccessForbidden, "no api access", errors.CatAPI, false, nil), time.Second)
	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{`"ok":false`, `"code":"API_ACCESS_FORBIDDEN"`, `"category":"api"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("error envelope missing %s: %s", want, got)
		}
	}
}

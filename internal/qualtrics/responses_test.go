package qualtrics

import (
	"context"
	"io"
	"mime"
	"net/http"
	"testing"

	"github.com/thedavidweng/qualtrics-cli/internal/testutil"
)

func TestGetExportParsesStatus(t *testing.T) {
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		return testutil.JSONResponse(200, fixture(t, "export_status.json")), nil
	})

	status, err := client.GetExport(context.Background(), "ES_abc")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != JobComplete || status.PercentComplete != 100 || status.FileID != "F_1" {
		t.Fatalf("status = %+v", status)
	}
}

func TestDownloadExportReturnsBytesAndFilename(t *testing.T) {
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		resp := testutil.JSONResponse(200, "PK\u0003\u0004binaryzipdata")
		resp.Header.Set("Content-Disposition", `attachment; filename="My Survey.zip"`)
		return resp, nil
	})

	data, filename, err := client.DownloadExport(context.Background(), "ES_abc")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "PK\u0003\u0004binaryzipdata" {
		t.Fatalf("data = %q", data)
	}
	if filename != "My Survey.zip" {
		t.Fatalf("filename = %s", filename)
	}
}

func TestUploadImportSendsMultipartFile(t *testing.T) {
	var gotPath, gotContentType string
	var gotFilename string
	var gotContent []byte
	client := stub(t, func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("not multipart: %v", err)
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		gotFilename = part.FileName()
		gotContent, _ = io.ReadAll(part)
		return testutil.JSONResponse(200, `{"result":null}`), nil
	})

	err := client.UploadImport(context.Background(), "ES_abc", []byte("ResponseId,Q1\nR_1,2\n"), "responses.csv")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/API/v3/responseimports/ES_abc/file" {
		t.Fatalf("path = %s", gotPath)
	}
	mediaType, _, err := mime.ParseMediaType(gotContentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type = %s", gotContentType)
	}
	if gotFilename != "responses.csv" {
		t.Fatalf("filename = %s", gotFilename)
	}
	if string(gotContent) != "ResponseId,Q1\nR_1,2\n" {
		t.Fatalf("content = %q", gotContent)
	}
}

package qualtrics

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
)

const maxDownloadBody = 200 << 20

type JobStatus string

const (
	JobInProgress JobStatus = "inProgress"
	JobComplete   JobStatus = "complete"
	JobFailed     JobStatus = "failed"
)

type ExportRequest struct {
	SurveyID             string   `json:"surveyId"`
	Format               string   `json:"format"`
	UseLabels            *bool    `json:"useLabels,omitempty"`
	TimeZone             string   `json:"timeZone,omitempty"`
	Compress             *bool    `json:"compress,omitempty"`
	BreakoutSets         []string `json:"breakoutSets,omitempty"`
	SeenUnansweredRecode string   `json:"seenUnansweredRecode,omitempty"`
}

type ExportStart struct {
	ProgressID string `json:"progressId"`
}

type ExportStatus struct {
	Status          JobStatus `json:"status"`
	PercentComplete float64   `json:"percentComplete"`
	FileID          string    `json:"fileId"`
}

func (c *Client) StartExport(ctx context.Context, req *ExportRequest) (ExportStart, error) {
	var out ExportStart
	if err := c.Do(ctx, "POST", "/responseexports", req, &out); err != nil {
		return ExportStart{}, err
	}
	return out, nil
}

func (c *Client) GetExport(ctx context.Context, progressID string) (ExportStatus, error) {
	var out ExportStatus
	if err := c.Do(ctx, "GET", "/responseexports/"+progressID, nil, &out); err != nil {
		return ExportStatus{}, err
	}
	return out, nil
}

func (c *Client) Download(ctx context.Context, path string) (data []byte, filename string, err error) {
	attempts := maxRetries + 1
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, backoff(attempt, nil)); err != nil {
				return nil, "", err
			}
		}
		req, err := c.newRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, "", err
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = errors.New(errors.NetworkUnreachable, "download failed", errors.CatNetwork, true, err)
			continue
		}
		raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxDownloadBody))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = errors.New(errors.NetworkUnreachable, "failed to read download", errors.CatNetwork, true, readErr)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr, retryAfter := mapStatus(resp.StatusCode, raw)
			if !apiErr.Retryable {
				return nil, "", apiErr
			}
			lastErr = apiErr
			if retryAfter > 0 {
				if err := sleepCtx(ctx, retryAfter); err != nil {
					return nil, "", err
				}
			}
			continue
		}
		return raw, filenameFromDisposition(resp.Header.Get("Content-Disposition")), nil
	}
	return nil, "", lastErr
}

func (c *Client) DownloadExport(ctx context.Context, progressID string) (data []byte, filename string, err error) {
	return c.Download(ctx, "/responseexports/"+progressID+"/file")
}

func filenameFromDisposition(value string) string {
	if value == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(value)
	if err != nil {
		return ""
	}
	return params["filename"]
}

type ImportStart struct {
	ProgressID string `json:"progressId"`
}

type ImportStatus struct {
	Status          JobStatus `json:"status"`
	PercentComplete float64   `json:"percentComplete"`
}

func (c *Client) StartImport(ctx context.Context, surveyID string) (ImportStart, error) {
	var out ImportStart
	req := map[string]string{"surveyId": surveyID}
	if err := c.Do(ctx, "POST", "/responseimports", req, &out); err != nil {
		return ImportStart{}, err
	}
	return out, nil
}

func (c *Client) GetImport(ctx context.Context, progressID string) (ImportStatus, error) {
	var out ImportStatus
	if err := c.Do(ctx, "GET", "/responseimports/"+progressID, nil, &out); err != nil {
		return ImportStatus{}, err
	}
	return out, nil
}

func (c *Client) UploadImport(ctx context.Context, progressID string, content []byte, filename string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return errors.New(errors.InternalError, "failed to build upload", errors.CatInternal, false, err)
	}
	if _, err := part.Write(content); err != nil {
		return errors.New(errors.InternalError, "failed to build upload", errors.CatInternal, false, err)
	}
	if err := w.Close(); err != nil {
		return errors.New(errors.InternalError, "failed to build upload", errors.CatInternal, false, err)
	}

	req, err := c.newRequest(ctx, "POST", "/responseimports/"+progressID+"/file", buf.Bytes())
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return errors.New(errors.NetworkUnreachable, "upload failed", errors.CatNetwork, false, err)
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr, _ := mapStatus(resp.StatusCode, raw)
		return apiErr
	}
	return nil
}

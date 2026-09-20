package qualtrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thedavidweng/qualtrics-cli/internal/errors"
	"github.com/thedavidweng/qualtrics-cli/internal/version"
)

const (
	maxRetries      = 3
	maxRetryWait    = 10 * time.Second
	maxResponseBody = 10 << 20
)

type Client struct {
	BaseURL   string
	Token     string
	UserAgent string
	HTTP      *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 8,
	}
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Token:     token,
		UserAgent: "qualtrics-cli/" + version.GetVersion(),
		HTTP: &http.Client{
			Transport: transport,
			Timeout:   timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *Client) TokenValue() string { return c.Token }

type apiErrorPayload struct {
	Meta struct {
		HTTPStatus string `json:"httpStatus"`
		RequestID  string `json:"requestId"`
		Error      struct {
			ErrorCode    string `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		} `json:"error"`
		Notice string `json:"notice"`
	} `json:"meta"`
}

type apiEnvelope struct {
	Result json.RawMessage `json:"result"`
	Meta   struct {
		HTTPStatus string `json:"httpStatus"`
		RequestID  string `json:"requestId"`
		Error      struct {
			ErrorCode    string `json:"errorCode"`
			ErrorMessage string `json:"errorMessage"`
		} `json:"error"`
	} `json:"meta"`
}

func (c *Client) Do(ctx context.Context, method, path string, body, result any) error {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return errors.New(errors.InternalError, "failed to encode request body", errors.CatInternal, false, err)
		}
	}

	attempts := maxRetries + 1
	if method == http.MethodPost {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := sleepCtx(ctx, backoff(attempt, nil)); err != nil {
				return err
			}
		}

		req, err := c.newRequest(ctx, method, path, payload)
		if err != nil {
			return err
		}

		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = errors.New(errors.NetworkUnreachable, "request to Qualtrics failed", errors.CatNetwork, true, err)
			continue
		}

		raw, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = errors.New(errors.NetworkUnreachable, "failed to read response", errors.CatNetwork, true, readErr)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return decodeSuccess(raw, result)
		}

		apiErr, retryAfter := mapStatus(resp.StatusCode, raw)
		if !apiErr.Retryable {
			return apiErr
		}
		lastErr = apiErr
		if retryAfter > 0 {
			if err := sleepCtx(ctx, retryAfter); err != nil {
				return err
			}
			continue
		}
	}
	return lastErr
}

func (c *Client) newRequest(ctx context.Context, method, path string, payload []byte) (*http.Request, error) {
	url := c.BaseURL + path
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, errors.New(errors.InternalError, "failed to build request", errors.CatInternal, false, err)
	}
	req.Header.Set("X-API-TOKEN", c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func decodeSuccess(raw []byte, result any) error {
	if result == nil || len(raw) == 0 {
		return nil
	}
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return errors.New(errors.APISchemaChanged, "failed to parse Qualtrics response", errors.CatAPI, false, err)
	}
	if len(env.Result) == 0 || string(env.Result) == "null" {
		return nil
	}
	if err := json.Unmarshal(env.Result, result); err != nil {
		return errors.New(errors.APISchemaChanged, "failed to parse Qualtrics result payload", errors.CatAPI, false, err)
	}
	return nil
}

func mapStatus(status int, raw []byte) (*errors.Error, time.Duration) {
	var payload apiErrorPayload
	_ = json.Unmarshal(raw, &payload)

	code := payload.Meta.Error.ErrorCode
	message := payload.Meta.Error.ErrorMessage
	requestID := payload.Meta.RequestID
	if message == "" {
		message = fmt.Sprintf("unexpected HTTP %d from Qualtrics", status)
	}
	if requestID != "" {
		message = fmt.Sprintf("%s (requestId: %s)", message, requestID)
	}

	switch {
	case status == http.StatusUnauthorized:
		return errors.New(errors.AuthTokenInvalid, "Qualtrics did not recognize the API token; run `qualtrics auth set-token`", errors.CatAuth, false, nil), 0
	case status == http.StatusForbidden && code == "AuthZ_2.0":
		return errors.New(errors.APIAccessForbidden, "token is valid but this user/brand has no access to the API endpoint; enable the API feature and the user's API permission in Qualtrics", errors.CatAPI, false, nil), 0
	case status == http.StatusTooManyRequests:
		return errors.NewWithRetryAfter(errors.RateLimited, message, errors.CatNetwork, true, parseRetryAfter(raw), nil), parseRetryAfter(raw)
	case status == http.StatusNotFound:
		return errors.New(errors.ResourceNotFound, message, errors.CatAPI, false, nil), 0
	case status == http.StatusBadRequest:
		return errors.New(errors.ValidationFailed, message, errors.CatValidation, false, nil), 0
	case status >= 500:
		return errors.New(errors.APIError, message, errors.CatAPI, true, nil), 0
	default:
		return errors.New(errors.APIError, message, errors.CatAPI, false, nil), 0
	}
}

func parseRetryAfter(raw []byte) time.Duration {
	var payload struct {
		Meta struct {
			Error struct {
				RetryAfter string `json:"retryAfter"`
			} `json:"error"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0
	}
	v := strings.TrimSpace(payload.Meta.Error.RetryAfter)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		d := time.Duration(secs) * time.Second
		if d > maxRetryWait {
			d = maxRetryWait
		}
		return d
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > maxRetryWait {
			d = maxRetryWait
		}
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}

func backoff(attempt int, _ *http.Response) time.Duration {
	d := time.Duration(1<<uint(attempt)) * 500 * time.Millisecond
	jitter := time.Duration(rand.Int64N(int64(500 * time.Millisecond)))
	d += jitter
	if d > maxRetryWait {
		d = maxRetryWait
	}
	return d
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return errors.New(errors.NetworkTimeout, "context canceled while waiting to retry", errors.CatNetwork, false, ctx.Err())
	case <-t.C:
		return nil
	}
}

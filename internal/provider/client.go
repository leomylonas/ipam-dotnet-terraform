package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type problemDetails struct {
	Type   string              `json:"type"`
	Title  string              `json:"title"`
	Status int                 `json:"status"`
	Detail string              `json:"detail"`
	Errors map[string][]string `json:"errors"`
}

type apiClient struct {
	baseURL      string
	username     string
	password     string
	client       *http.Client
	maxRetries   int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
}

func newAPIClient(baseURL, username, password string, timeout time.Duration, insecureSkipTLS bool, maxRetries int, retryWaitMin, retryWaitMax time.Duration) (*apiClient, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("base_url must be an absolute URL")
	}

	transport := &http.Transport{}
	if insecureSkipTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &apiClient{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		maxRetries:   maxRetries,
		retryWaitMin: retryWaitMin,
		retryWaitMax: retryWaitMax,
	}, nil
}

func (c *apiClient) doJSON(ctx context.Context, method, path string, reqBody any, out any, expected ...int) error {
	var payload []byte
	var err error
	if reqBody != nil {
		payload, err = json.Marshal(reqBody)
		if err != nil {
			return err
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var body io.Reader
		if reqBody != nil {
			body = bytes.NewReader(payload)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return err
		}
		if reqBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.SetBasicAuth(c.username, c.password)

		resp, err := c.client.Do(req)
		if err != nil {
			if isRetryableNetErr(err) && attempt < c.maxRetries {
				sleepWithBackoff(c.retryWaitMin, c.retryWaitMax, attempt)
				continue
			}
			return err
		}

		respBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if isExpected(resp.StatusCode, expected...) {
			if out != nil && len(respBytes) > 0 {
				if err := json.Unmarshal(respBytes, out); err != nil {
					return err
				}
			}
			return nil
		}

		if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) && attempt < c.maxRetries {
			sleepWithBackoff(c.retryWaitMin, c.retryWaitMax, attempt)
			continue
		}

		lastErr = newAPIError(resp.StatusCode, respBytes, resp.Header.Get("Content-Type"))
		break
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("request failed without response")
}

func sleepWithBackoff(minWait, maxWait time.Duration, attempt int) {
	wait := minWait * time.Duration(1<<attempt)
	if wait > maxWait {
		wait = maxWait
	}
	time.Sleep(wait)
}

func isExpected(code int, expected ...int) bool {
	for _, e := range expected {
		if code == e {
			return true
		}
	}
	return false
}

func isRetryableNetErr(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}

type apiError struct {
	StatusCode int
	Body       string
	Title      string
	Detail     string
	Validation map[string][]string
}

func (e *apiError) Error() string {
	base := fmt.Sprintf("ipam api returned status %d", e.StatusCode)

	if e.Title != "" && e.Detail != "" {
		base = fmt.Sprintf("%s: %s - %s", base, e.Title, e.Detail)
	} else if e.Title != "" {
		base = fmt.Sprintf("%s: %s", base, e.Title)
	} else if e.Detail != "" {
		base = fmt.Sprintf("%s: %s", base, e.Detail)
	}

	if len(e.Validation) > 0 {
		pairs := make([]string, 0)
		for field, errs := range e.Validation {
			if len(errs) == 0 {
				continue
			}
			pairs = append(pairs, fmt.Sprintf("%s=%s", field, strings.Join(errs, "; ")))
		}
		if len(pairs) > 0 {
			return fmt.Sprintf("%s (validation: %s)", base, strings.Join(pairs, ", "))
		}
	}

	if e.Body != "" && e.Title == "" && e.Detail == "" {
		return fmt.Sprintf("%s: %s", base, e.Body)
	}
	return base
}

func newAPIError(status int, body []byte, contentType string) *apiError {
	err := &apiError{
		StatusCode: status,
		Body:       strings.TrimSpace(string(body)),
	}

	if len(body) == 0 {
		return err
	}

	// ASP.NET ProblemDetails typically use application/problem+json but some
	// middlewares may still return application/json; parse both.
	if strings.Contains(contentType, "json") {
		var pd problemDetails
		if json.Unmarshal(body, &pd) == nil {
			if pd.Title != "" || pd.Detail != "" || len(pd.Errors) > 0 {
				err.Title = pd.Title
				err.Detail = pd.Detail
				err.Validation = pd.Errors
			}
		}
	}

	return err
}

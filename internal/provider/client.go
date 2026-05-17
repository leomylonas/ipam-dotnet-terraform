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

type apiClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

func newAPIClient(baseURL, username, password string, timeout time.Duration, insecureSkipTLS bool) (*apiClient, error) {
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
	for attempt := 0; attempt < 3; attempt++ {
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
			if isRetryableNetErr(err) && attempt < 2 {
				time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
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

		if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) && attempt < 2 {
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}

		lastErr = &apiError{StatusCode: resp.StatusCode, Body: string(respBytes)}
		break
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("request failed without response")
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
}

func (e *apiError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("ipam api returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("ipam api returned status %d: %s", e.StatusCode, e.Body)
}

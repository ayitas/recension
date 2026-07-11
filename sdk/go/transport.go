package recension

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const sdkVersion = "0.1.0"

type transport struct {
	apiKey string
	apiURL string
	client *http.Client
}

func newTransport() *transport {
	return &transport{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (t *transport) configure(opts Options) error {
	t.apiKey = opts.APIKey
	t.apiURL = strings.TrimRight(opts.APIURL, "/")
	body, _ := json.Marshal(map[string]string{"team": opts.Team})
	req, err := http.NewRequest(http.MethodPost, t.apiURL+"/v1/client/verify", bytes.NewReader(body))
	if err != nil {
		return err
	}
	t.setHeaders(req, "application/json")
	res, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("recension: server unreachable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("recension: invalid api key")
	}
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("recension: verify failed: status=%d body=%s", res.StatusCode, string(b))
	}
	return nil
}

func (t *transport) post(path string, body []byte) ([]byte, error) {
	var rdr io.Reader
	contentType := "application/json"
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, t.apiURL+path, rdr)
	if err != nil {
		return nil, err
	}
	t.setHeaders(req, contentType)
	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}
	if res.StatusCode == http.StatusOK {
		return b, nil
	}
	return nil, fmt.Errorf("recension: post %s failed: status=%d body=%s", path, res.StatusCode, string(b))
}

func (t *transport) head(path string) (bool, error) {
	req, err := http.NewRequest(http.MethodHead, t.apiURL+path, nil)
	if err != nil {
		return false, err
	}
	t.setHeaders(req, "")
	res, err := t.client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusNoContent, http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		b, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("recension: head %s failed: status=%d body=%s", path, res.StatusCode, string(b))
	}
}

func (t *transport) putRaw(path string, body []byte, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req, err := http.NewRequest(http.MethodPut, t.apiURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	t.setHeaders(req, contentType)
	res, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent {
		return nil
	}
	b, _ := io.ReadAll(res.Body)
	return fmt.Errorf("recension: put %s failed: status=%d body=%s", path, res.StatusCode, string(b))
}

func (t *transport) setHeaders(req *http.Request, contentType string) {
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("User-Agent", "recension-go/"+sdkVersion)
	req.Header.Set("X-Recension-API-Key", t.apiKey)
}

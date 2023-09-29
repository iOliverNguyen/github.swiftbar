package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func httpGET(url string) ([]byte, error) {
	return httpRequest("GET", url, nil)
}

func httpPOST(url string, body any) ([]byte, error) {
	return httpRequest("POST", url, body)
}

func httpRequest(method string, url string, body any) (_ []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var bodyReader io.Reader
	var bodyJSON []byte
	if body != nil {
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+githubToken)

	debugf("-> %v %v", method, url)
	if bodyJSON != nil {
		debugf("%s\n", bodyJSON)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		debugf("failed to call http request: %v %v", url, err)
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		debugf("<- %v", resp.Status)
		// debugf("%s\n", data)
		return data, err
	}
	debugf("failed to call http request: %v %v", url, resp.Status)
	debugf("%s", data)
	return data, errorf("failed to call http request: (%v) %s", resp.Status, data)
}

func htmlRequest(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, s := range htmlHeaders {
		idx := strings.Index(s, ":")
		key, value := s[:idx], strings.TrimSpace(s[idx+1:])
		req.Header.Set(key, value)
	}

	debugf("-> HTML %v", url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		debugf("failed to call html request: %v %v", url, err)
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		debugf("<- %v", resp.Status)
		return data, nil
	}
	debugf("failed to call html request: %v %v", url, resp.Status)
	debugf("%s", data)
	return data, errorf("failed to call html request: %v (%v) %s", url, resp.Status, data)
}

package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type MockHttpClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestRetryRequest(t *testing.T) {
	mockClient := &MockHttpClient{}
	client := &retryClient{client: mockClient}
	url := "http://example.com"

	t.Run("TestMaxRetriesExceeded", func(t *testing.T) {
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == url {
				return &http.Response{StatusCode: http.StatusUnauthorized}, nil
			}
			return nil, errors.New("unknown URL")
		}

		_, err := client.retryRequest(url, 1)

		if err == nil || !strings.Contains(err.Error(), "max retries exceeded") {
			t.Errorf("Expected 'max retries exceeded' error, got %v", err)
		}
	})

	t.Run("TestSuccessfulRequest", func(t *testing.T) {
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == url {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("success"))}, nil
			}
			return nil, errors.New("unknown URL")
		}

		body, err := client.retryRequest(url, 1)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if string(body) != "success" {
			t.Errorf("Expected body to be 'success', got %s", body)
		}
	})

	t.Run("TestNon401Failure", func(t *testing.T) {
		mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == url {
				return &http.Response{StatusCode: http.StatusInternalServerError}, nil
			}
			return nil, errors.New("unknown URL")
		}

		_, err := client.retryRequest(url, 1)

		if err == nil || !strings.Contains(err.Error(), "received non-200 response") {
			t.Errorf("Expected 'received non-200 response' error, got %v", err)
		}
	})
}

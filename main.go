package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type retryClient struct {
	client HttpClient
}

func main() {
	fmt.Println("Hello, World!")

	url := "http://example.com"
	client := &retryClient{
		client: &http.Client{
			Timeout: time.Second * 20,
		},
	}
	body, err := client.retryRequest(url, 5)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(body))
	}
}

func (rc *retryClient) retryRequest(url string, maxRetries int) ([]byte, error) {
	var body []byte
	retryCount := 0
	minBackoffTime := time.Second
	maxBackoffTime := time.Second * 2

	for retryCount <= maxRetries {
		client := http.Client{
			Timeout: time.Second * 20,
		}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusUnauthorized {
			// we need to retry after a random little wait
			randomBackoff := time.Duration(rand.Int63n(int64(maxBackoffTime-minBackoffTime))) + minBackoffTime
			time.Sleep(randomBackoff)

			minBackoffTime *= 2
			maxBackoffTime *= 2

			retryCount++
		} else if resp.StatusCode == http.StatusOK {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			err = resp.Body.Close()
			if err != nil {
				return nil, err
			}

			break // we're done, no need to retry
		} else {
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			err = resp.Body.Close()
			if err != nil {
				return nil, err
			}

			return body, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

	}
	if retryCount > maxRetries {
		return nil, fmt.Errorf("max retries exceeded")
	}
	return body, nil
}

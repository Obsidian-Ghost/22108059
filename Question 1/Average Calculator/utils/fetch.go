package utils

import (
	"Average_Calculator/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var apiEndpoints = map[string]string{
	"p": "http://20.244.56.144/evaluation-service/primes",
	"f": "http://20.244.56.144/evaluation-service/fibo",
	"e": "http://20.244.56.144/evaluation-service/evens",
	"r": "http://20.244.56.144/evaluation-service/rand",
}

const (
	WindowSize = 10
	TimeoutMS  = 500
	Port       = ":9876"
)

func FetchNumbers(numberType string) ([]int, error) {
	endpoint, exists := apiEndpoints[numberType]
	if !exists {
		return nil, fmt.Errorf("invalid number type: %s", numberType)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(TimeoutMS) * time.Millisecond,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(TimeoutMS)*time.Millisecond)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response time
	responseTime := time.Since(startTime)
	if responseTime > time.Duration(TimeoutMS)*time.Millisecond {
		return nil, fmt.Errorf("response time exceeded %dms", TimeoutMS)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var numbersResp models.NumbersResponse
	if err := json.Unmarshal(body, &numbersResp); err != nil {
		return nil, err
	}

	return numbersResp.Numbers, nil
}

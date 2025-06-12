package main

import (
	"Average_Calculator/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	WindowSize = 10
	TimeoutMS  = 500
	Port       = ":9876"
)

// in memory db
var numberWindow []int

// provided api endpoints
var apiEndpoints = map[string]string{
	"p": "http://20.244.56.144/evaluation-service/primes",
	"f": "http://20.244.56.144/evaluation-service/fibo",
	"e": "http://20.244.56.144/evaluation-service/evens",
	"r": "http://20.244.56.144/evaluation-service/rand",
}

func fetchNumbers(numberType string) ([]int, error) {
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

func addUniqueNumbers(newNumbers []int) []int {
	prevState := make([]int, len(numberWindow))
	copy(prevState, numberWindow)

	for _, num := range newNumbers {
		if !contains(numberWindow, num) {
			numberWindow = append(numberWindow, num)
		}
	}

	if len(numberWindow) > WindowSize {
		excess := len(numberWindow) - WindowSize
		numberWindow = numberWindow[excess:]
	}

	return prevState
}

func contains(slice []int, num int) bool {
	for _, item := range slice {
		if item == num {
			return true
		}
	}
	return false
}

func calculateAverage() float64 {
	if len(numberWindow) == 0 {
		return 0.0
	}

	sum := 0
	for _, num := range numberWindow {
		sum += num
	}

	return float64(sum) / float64(len(numberWindow))
}

func numbersHandler(c echo.Context) error {
	numberType := c.Param("numberid")

	if _, exists := apiEndpoints[numberType]; !exists {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid type",
		})
	}

	newNumbers, err := fetchNumbers(numberType)
	if err != nil {
		fmt.Printf("Error fetching numbers for type '%s': %v\n", numberType, err)
		newNumbers = []int{}
	}

	prevState := addUniqueNumbers(newNumbers)

	avg := calculateAverage()

	response := models.APIResponse{
		WindowPrevState: prevState,
		WindowCurrState: make([]int, len(numberWindow)),
		Numbers:         newNumbers,
		Avg:             avg,
	}

	return c.JSON(http.StatusOK, response)
}

func main() {
	//init echo instanse
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Add timeout middleware for all requests
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(TimeoutMS) * time.Millisecond,
	}))

	RoutesInit(*e)

	// Start server
	fmt.Printf("server starting on port %s\n", Port)

	e.Logger.Fatal(e.Start(Port))
}

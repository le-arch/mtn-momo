package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

const (
	defaultBaseURL      = "https://demo.campay.net/api"
	pollInterval        = 3 * time.Second
	maxPollAttempts     = 40
	requestTimeout      = 10 * time.Second
	minPhoneLength      = 9
	minAmount           = 0.0
)

// Config holds application configuration
type Config struct {
	BaseURL         string
	APIKey          string
	PollInterval    time.Duration
	MaxPollAttempts int
}

// PaymentRequest represents a payment collection request
type PaymentRequest struct {
	Amount      string `json:"amount"`
	From        string `json:"from"`
	Description string `json:"description"`
}

// InitResponse represents the API response for payment initiation
type InitResponse struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// StatusResponse represents the API response for status checks
type StatusResponse struct {
	Status string `json:"status"`
}

// CampayClient handles all interactions with the Campay API
type CampayClient struct {
	client *resty.Client
	config *Config
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create Campay client
	client := NewCampayClient(config)

	// Get user input
	paymentReq, err := getUserInput()
	if err != nil {
		return fmt.Errorf("failed to get user input: %w", err)
	}

	// Validate input
	if err := validatePaymentRequest(paymentReq); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	fmt.Println("\n=== Payment Details ===")
	fmt.Printf("Number: %s\nAmount: %s\nDescription: %s\n", 
		paymentReq.From, paymentReq.Amount, paymentReq.Description)
	fmt.Println("\nSending payment request to Campay...")

	// Initiate payment
	reference, err := client.InitiatePayment(paymentReq)
	if err != nil {
		return fmt.Errorf("failed to initiate payment: %w", err)
	}

	fmt.Printf("\n✓ Transaction initialized\nReference: %s\n", reference)
	fmt.Println("Waiting for MTN Mobile Money confirmation...")

	// Poll for transaction status with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 
		time.Duration(config.MaxPollAttempts)*config.PollInterval)
	defer cancel()

	status, err := client.PollTransactionStatus(ctx, reference)
	if err != nil {
		return fmt.Errorf("failed to get transaction status: %w", err)
	}

	fmt.Println("\n=== FINAL TRANSACTION STATUS ===")
	fmt.Println(status)

	return nil
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	// Load .env file if it exists (ignore error if not found)
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return nil, fmt.Errorf("failed to load .env: %w", err)
		}
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is required")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Config{
		BaseURL:         baseURL,
		APIKey:          apiKey,
		PollInterval:    pollInterval,
		MaxPollAttempts: maxPollAttempts,
	}, nil
}

// NewCampayClient creates a new Campay API client
func NewCampayClient(config *Config) *CampayClient {
	client := resty.New().
		SetTimeout(requestTimeout).
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second).
		SetHeader("Authorization", "Token "+config.APIKey).
		SetHeader("Content-Type", "application/json")

	return &CampayClient{
		client: client,
		config: config,
	}
}

// getUserInput prompts the user for payment details
func getUserInput() (*PaymentRequest, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter mobile money number: ")
	momoNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read mobile number: %w", err)
	}
	momoNumber = strings.TrimSpace(momoNumber)

	fmt.Print("Enter amount: ")
	amount, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read amount: %w", err)
	}
	amount = strings.TrimSpace(amount)

	fmt.Print("Enter description: ")
	description, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read description: %w", err)
	}
	description = strings.TrimSpace(description)

	return &PaymentRequest{
		Amount:      amount,
		From:        momoNumber,
		Description: description,
	}, nil
}

// validatePaymentRequest validates the payment request fields
func validatePaymentRequest(req *PaymentRequest) error {
	if err := validatePhoneNumber(req.From); err != nil {
		return err
	}
	if err := validateAmount(req.Amount); err != nil {
		return err
	}
	if err := validateDescription(req.Description); err != nil {
		return err
	}
	return nil
}

// validatePhoneNumber validates Cameroon phone numbers
func validatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Remove common formatting characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "+", "")

	if len(phone) < minPhoneLength {
		return fmt.Errorf("phone number too short (minimum %d digits)", minPhoneLength)
	}

	// Basic validation for Cameroon numbers (starting with 237 or 6)
	if !strings.HasPrefix(phone, "237") && !strings.HasPrefix(phone, "6") {
		return fmt.Errorf("invalid phone number format (should start with 237 or 6)")
	}

	return nil
}

// validateAmount validates the payment amount
func validateAmount(amount string) error {
	if amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}

	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount format: must be a number")
	}

	if val <= minAmount {
		return fmt.Errorf("amount must be greater than %.2f", minAmount)
	}

	return nil
}

// validateDescription validates the payment description
func validateDescription(description string) error {
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if len(description) > 200 {
		return fmt.Errorf("description too long (maximum 200 characters)")
	}
	return nil
}

// InitiatePayment sends a payment collection request to Campay
func (c *CampayClient) InitiatePayment(req *PaymentRequest) (string, error) {
	var initResp InitResponse

	resp, err := c.client.R().
		SetBody(req).
		SetResult(&initResp).
		Post(c.config.BaseURL + "/collect/")

	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return "", fmt.Errorf("API error (status %d): %s", 
			resp.StatusCode(), initResp.Message)
	}

	if initResp.Reference == "" {
		return "", fmt.Errorf("no reference returned: %s", initResp.Message)
	}

	return initResp.Reference, nil
}

// PollTransactionStatus polls the transaction status until completion or timeout
func (c *CampayClient) PollTransactionStatus(ctx context.Context, reference string) (string, error) {
	attempts := 0

	for attempts < c.config.MaxPollAttempts {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("transaction status check cancelled: %w", ctx.Err())
		default:
		}

		status, err := c.CheckTransactionStatus(reference)
		if err != nil {
			return "", fmt.Errorf("failed to check status: %w", err)
		}

		switch status {
		case "SUCCESSFUL":
			return "✓ Transaction Successful", nil
		case "FAILED":
			return "✗ Transaction Failed", nil
		}

		attempts++
		fmt.Printf("Status: PENDING... (attempt %d/%d)\n", attempts, c.config.MaxPollAttempts)

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("transaction timeout: %w", ctx.Err())
		case <-time.After(c.config.PollInterval):
			// Continue polling
		}
	}

	return "", fmt.Errorf("transaction timeout after %d attempts", c.config.MaxPollAttempts)
}

// CheckTransactionStatus checks the current status of a transaction
func (c *CampayClient) CheckTransactionStatus(reference string) (string, error) {
	var statusResp StatusResponse

	resp, err := c.client.R().
		SetResult(&statusResp).
		Get(c.config.BaseURL + "/transaction/" + reference + "/")

	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return "", fmt.Errorf("API error (status %d): %s", 
			resp.StatusCode(), resp.String())
	}

	return statusResp.Status, nil
}
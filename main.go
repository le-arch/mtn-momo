package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

const (
	baseURL = "https://demo.campay.net/api"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func run() error {
	// Load .env if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("failed to load .env: %w", err)
		}
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return fmt.Errorf("API_KEY is not set in .env")
	}

	client := resty.New()

	reader := bufio.NewReader(os.Stdin)

	// Prompt user input 
	// prompt for mobile money number
	fmt.Print("Enter mobile money number: ")
	momoNumber, _ := reader.ReadString('\n')
	momoNumber = strings.TrimSpace(momoNumber)
	
	// prompt for amount
	fmt.Print("Enter amount: ")
	amount, _ := reader.ReadString('\n')
	amount = strings.TrimSpace(amount)
	
	// prompt for description
	fmt.Print("Enter description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Println("\nSending payment request to Campay...")
	fmt.Printf("Number: %s\nAmount: %s\nDescription: %s\n", momoNumber, amount, description)

	// STEP 1: Initiate transaction 
	initResp := struct {
		Reference string `json:"reference"`
		Status    string `json:"status"`
		Message   string `json:"message"`
	}{}

	_, err := client.R().
		SetHeader("Authorization", "Token "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"amount":      amount,
			"from":        momoNumber,
			"description": description,
		}).
		SetResult(&initResp).
		Post(baseURL + "/collect/")

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if initResp.Reference == "" {
		return fmt.Errorf("failed to start transaction: %s", initResp.Message)
	}

	fmt.Println("Transaction initialized. Waiting for MTN Mobile Money confirmation...")
	fmt.Println("Reference:", initResp.Reference)

	// STEP 2: Poll the transaction status 
	finalStatus, err := pollTransactionStatus(client, apiKey, initResp.Reference)
	if err != nil {
		return err
	}

	fmt.Println("\n=== FINAL TRANSACTION STATUS ===")
	fmt.Println(finalStatus)

	return nil
}

func pollTransactionStatus(client *resty.Client, apiKey, ref string) (string, error) {
	statusResp := struct {
		Status string `json:"status"`
	}{}

	for {
		_, err := client.R().
			SetHeader("Authorization", "Token "+apiKey).
			SetResult(&statusResp).
			Get(baseURL + "/transaction/" + ref + "/")

		if err != nil {
			return "", fmt.Errorf("failed to check transaction status: %w", err)
		}

		// Possible statuses: PENDING, FAILED, SUCCESSFUL
		switch statusResp.Status {
		case "SUCCESSFUL":
			return "Transaction Successful ✓", nil
		case "FAILED":
			return "Transaction Failed ✗", nil
		}

		fmt.Println("Still pending... waiting 3 seconds...")
		time.Sleep(3 * time.Second)
	}
}


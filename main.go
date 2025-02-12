/*
Author: Noah Caulfield
Date: 1/25/25
Description:
This is a simple receipt processing web service implemented in Go.
It allows users to submit receipts, calculates reward points based on predefined rules,
and retrieves the points awarded for a given receipt.
The application runs as a REST API and stores data in memory.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Receipt represents the structure of a receipt submitted by the user.
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []Item `json:"items"`
}

// Item represents an individual item on a receipt.
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// In-memory storage for receipt IDs and their corresponding points.
var receipts = make(map[string]int)

// validateReceipt ensures that the receipt structure is correct and contains valid data.
func validateReceipt(receipt Receipt) error {
	if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" || len(receipt.Items) == 0 {
		return fmt.Errorf("The receipt is invalid.") // Matches OpenAPI error response
	}
	if !regexp.MustCompile(`^[\w\s\-&]+$`).MatchString(receipt.Retailer) {
		return fmt.Errorf("The receipt is invalid: retailer name format is incorrect.")
	}
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return fmt.Errorf("The receipt is invalid: purchaseDate format must be YYYY-MM-DD.")
	}
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return fmt.Errorf("The receipt is invalid: purchaseTime format must be HH:MM (24-hour format).")
	}
	if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(receipt.Total) {
		return fmt.Errorf("The receipt is invalid: total format must be a decimal with two places.")
	}
	return nil
}

// calculatePoints applies predefined rules to determine the number of points a receipt earns.
func calculatePoints(receipt Receipt) int {
	points := 0
	alphanumericRegex := regexp.MustCompile("[a-zA-Z0-9]")

	// Rule 1: One point per alphanumeric character in the retailer name.
	retailerPoints := len(alphanumericRegex.FindAllString(receipt.Retailer, -1))
	points += retailerPoints

	total, _ := strconv.ParseFloat(receipt.Total, 64)

	// Rule 2: 50 points if the total is a round dollar amount with no cents.
	if math.Mod(total, 1) == 0 {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25.
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt.
	itemPairsPoints := (len(receipt.Items) / 2) * 5
	points += itemPairsPoints

	// Rule 5: Additional points if item description length is a multiple of 3.
	for _, item := range receipt.Items {
		desc := strings.TrimSpace(item.ShortDescription)
		price, _ := strconv.ParseFloat(item.Price, 64)
		descLen := len(desc)

		if descLen%3 == 0 {
			extraPoints := int(math.Ceil(price*0.2 + 0.0001)) // Fix rounding issue
			points += extraPoints
		}
	}

	// Rule 6: 6 points if the purchase day is odd.
	dateParts := strings.Split(receipt.PurchaseDate, "-")
	if len(dateParts) == 3 {
		day, _ := strconv.Atoi(dateParts[2])
		if day%2 == 1 {
			points += 6
		}
	}

	// Rule 7: 10 points if the purchase time is between 2:00 PM and 4:00 PM.
	t, _ := time.Parse("15:04", receipt.PurchaseTime)
	if t.Hour() >= 14 && t.Hour() < 16 {
		points += 10
	}

	// Rule 8: 5 points if the total is greater than 10.00.
	if total > 10.00 {
		points += 5
	}

	return points
}

// ProcessReceipt handles POST /receipts/process - Stores receipt and calculates points.
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil || validateReceipt(receipt) != nil {
		http.Error(w, "The receipt is invalid.", http.StatusBadRequest) // Matches OpenAPI
		return
	}
	receiptID := uuid.New().String()
	receipts[receiptID] = calculatePoints(receipt)
	response := map[string]string{"id": receiptID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPoints handles GET /receipts/{id}/points - Retrieves points for a given receipt ID.
func GetPoints(w http.ResponseWriter, r *http.Request) {
	receiptID := strings.TrimPrefix(r.URL.Path, "/receipts/")
	receiptID = strings.TrimSuffix(receiptID, "/points")
	points, exists := receipts[receiptID]
	if !exists {
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound) // Matches OpenAPI
		return
	}
	response := map[string]int{"points": points}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RootHandler provides instructions on how to use the API.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Receipt Processor API is running!\n")

	fmt.Fprintln(w, "Usage Instructions:\n")

	fmt.Fprintln(w, "1. Submit a receipt:\n")
	fmt.Fprintln(w, `   curl -X POST http://localhost:8080/receipts/process -H "Content-Type: application/json" -d '{
       "retailer": "Target",
       "purchaseDate": "2022-01-01",
       "purchaseTime": "13:01",
       "total": "35.35",
       "items": [
           {"shortDescription": "Mountain Dew 12PK", "price": "6.49"}
       ]
   }'`)
	fmt.Fprintln(w, "\n")
	fmt.Fprintln(w, "\n")
	fmt.Fprintln(w, "2. Retrieve receipt points:\n")
	fmt.Fprintln(w, "   curl -X GET http://localhost:8080/receipts/{id}/points")
	fmt.Fprintln(w, "   (Replace {id} with the actual receipt ID from the previous response)\n")
}

func main() {
	// Register API routes
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/receipts/process", ProcessReceipt)
	http.HandleFunc("/receipts/", GetPoints)

	// Start server
	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

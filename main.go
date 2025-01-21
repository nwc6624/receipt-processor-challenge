/*
Author: Noah Caulfield
Date: 1/24/25
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
	Retailer     string `json:"retailer"`     // Name of the retailer/store.
	PurchaseDate string `json:"purchaseDate"` // Date of purchase in YYYY-MM-DD format.
	PurchaseTime string `json:"purchaseTime"` // Time of purchase in HH:MM format (24-hour time).
	Total        string `json:"total"`        // Total amount paid for the receipt.
	Items        []Item `json:"items"`        // List of purchased items.
}

// Item represents an individual item on a receipt.
type Item struct {
	ShortDescription string `json:"shortDescription"` // Short description of the item.
	Price            string `json:"price"`            // Price of the item.
}

// In-memory storage for receipt IDs and their corresponding points.
var receipts = make(map[string]int)

// validateReceipt ensures that the receipt structure is correct and contains valid data.
func validateReceipt(receipt Receipt) error {
	if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" || len(receipt.Items) == 0 {
		return fmt.Errorf("Invalid receipt: missing required fields")
	}
	if !regexp.MustCompile(`^[\w\s\-&]+$`).MatchString(receipt.Retailer) {
		return fmt.Errorf("Invalid retailer name format")
	}
	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return fmt.Errorf("Invalid purchaseDate format")
	}
	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return fmt.Errorf("Invalid purchaseTime format")
	}
	if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(receipt.Total) {
		return fmt.Errorf("Invalid total format")
	}
	return nil
}

// calculatePoints applies predefined rules to determine the number of points a receipt earns.
func calculatePoints(receipt Receipt) int {
	points := 0
	reg := regexp.MustCompile("[a-zA-Z0-9]")

	// Rule: One point per alphanumeric character in the retailer name.
	retailerPoints := len(reg.FindAllString(receipt.Retailer, -1))
	fmt.Printf("Retailer Points: %d\n", retailerPoints)
	points += retailerPoints

	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if total == math.Floor(total) {
		fmt.Println("Round dollar amount: +50 points")
		points += 50
	}
	if math.Mod(total, 0.25) == 0 {
		fmt.Println("Multiple of 0.25: +25 points")
		points += 25
	}

	// Rule: 5 points for every two items on the receipt.
	itemPairsPoints := (len(receipt.Items) / 2) * 5
	fmt.Printf("Item Pairs Points: %d\n", itemPairsPoints)
	points += itemPairsPoints

	// Rule: Additional points if item description length is a multiple of 3.
	for _, item := range receipt.Items {
		desc := strings.TrimSpace(item.ShortDescription)
		price, _ := strconv.ParseFloat(item.Price, 64)
		if len(desc)%3 == 0 {
			extraPoints := int(math.Ceil(price * 0.2))
			fmt.Printf("Item '%s' (length %d) adds %d points\n", desc, len(desc), extraPoints)
			points += extraPoints
		}
	}

	// Rule: 6 points if the purchase day is odd.
	dateParts := strings.Split(receipt.PurchaseDate, "-")
	if len(dateParts) == 3 {
		day, _ := strconv.Atoi(dateParts[2])
		if day%2 == 1 {
			fmt.Println("Odd purchase day: +6 points")
			points += 6
		}
	}

	// Rule: 10 points if the purchase time is between 2:00 PM and 4:00 PM.
	t, _ := time.Parse("15:04", receipt.PurchaseTime)
	if t.Hour() >= 14 && t.Hour() < 16 {
		fmt.Println("Purchase time 2-4PM: +10 points")
		points += 10
	}

	fmt.Printf("Total calculated points: %d\n", points)
	return points
}

// ProcessReceipt handles POST /receipts/process - Stores receipt and calculates points.
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil || validateReceipt(receipt) != nil {
		http.Error(w, "Invalid receipt format. Please verify input.", http.StatusBadRequest)
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
		http.Error(w, "No receipt found for that ID.", http.StatusNotFound)
		return
	}
	response := map[string]int{"points": points}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RootHandler provides instructions on how to use the API.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Receipt Processor API is running!\n")
	fmt.Fprintln(w, "Usage Instructions:")
	fmt.Fprintln(w, "1. Submit a receipt:")
	fmt.Fprintln(w, `   curl -X POST http://localhost:8080/receipts/process -H "Content-Type: application/json" -d '{"retailer":"Target", "purchaseDate":"2022-01-01", "purchaseTime":"13:01", "total":"35.35", "items":[{"shortDescription":"Mountain Dew 12PK", "price":"6.49"}]}'`)
	fmt.Fprintln(w, "2. Retrieve receipt points:")
	fmt.Fprintln(w, "   curl -X GET http://localhost:8080/receipts/{id}/points")
	fmt.Fprintln(w, "   (Replace {id} with the actual receipt ID from the previous response)")
}

func main() {
	http.HandleFunc("/", RootHandler)
	http.HandleFunc("/receipts/process", ProcessReceipt)
	http.HandleFunc("/receipts/", GetPoints)
	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

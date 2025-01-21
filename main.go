
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"

    "github.com/google/uuid"
)

type Receipt struct {
    Retailer     string `json:"retailer"`
    PurchaseDate string `json:"purchaseDate"`
    PurchaseTime string `json:"purchaseTime"`
    Total        string `json:"total"`
    Items        []Item `json:"items"`
}

type Item struct {
    ShortDescription string `json:"shortDescription"`
    Price            string `json:"price"`
}

var receipts = make(map[string]int)

func calculatePoints(receipt Receipt) int {
    return 100
}

func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
    var receipt Receipt
    if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    receiptID := uuid.New().String()
    receipts[receiptID] = calculatePoints(receipt)

    json.NewEncoder(w).Encode(map[string]string{"id": receiptID})
}

func GetPoints(w http.ResponseWriter, r *http.Request) {
    receiptID := strings.TrimPrefix(r.URL.Path, "/receipts/")
    receiptID = strings.TrimSuffix(receiptID, "/points")

    points, exists := receipts[receiptID]
    if !exists {
        http.Error(w, "Receipt not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(map[string]int{"points": points})
}

func main() {
    http.HandleFunc("/receipts/process", ProcessReceipt)
    http.HandleFunc("/receipts/", GetPoints)

    fmt.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

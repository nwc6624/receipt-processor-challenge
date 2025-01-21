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

// Receipt structure
type Receipt struct {
    Retailer     string `json:"retailer"`
    PurchaseDate string `json:"purchaseDate"`
    PurchaseTime string `json:"purchaseTime"`
    Total        string `json:"total"`
    Items        []Item `json:"items"`
}

// Item structure
type Item struct {
    ShortDescription string `json:"shortDescription"`
    Price            string `json:"price"`
}

// In-memory storage
var receipts = make(map[string]int)

// Validate receipt structure
func validateReceipt(receipt Receipt) error {
    if receipt.Retailer == "" || receipt.PurchaseDate == "" || receipt.PurchaseTime == "" || receipt.Total == "" || len(receipt.Items) == 0 {
        return fmt.Errorf("Invalid receipt: missing required fields")
    }
    if !regexp.MustCompile(`^[\\w\\s\\-&]+$`).MatchString(receipt.Retailer) {
        return fmt.Errorf("Invalid retailer name format")
    }
    if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
        return fmt.Errorf("Invalid purchaseDate format")
    }
    if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
        return fmt.Errorf("Invalid purchaseTime format")
    }
    if !regexp.MustCompile(`^\\d+\\.\\d{2}$`).MatchString(receipt.Total) {
        return fmt.Errorf("Invalid total format")
    }
    return nil
}

// Calculate points based on the rules
func calculatePoints(receipt Receipt) int {
    points := 0
    reg := regexp.MustCompile("[a-zA-Z0-9]")
    points += len(reg.FindAllString(receipt.Retailer, -1))
    total, _ := strconv.ParseFloat(receipt.Total, 64)
    if total == math.Floor(total) {
        points += 50
    }
    if math.Mod(total, 0.25) == 0 {
        points += 25
    }
    points += (len(receipt.Items) / 2) * 5
    for _, item := range receipt.Items {
        desc := strings.TrimSpace(item.ShortDescription)
        price, _ := strconv.ParseFloat(item.Price, 64)
        if len(desc)%3 == 0 {
            points += int(math.Ceil(price * 0.2))
        }
    }
    dateParts := strings.Split(receipt.PurchaseDate, "-")
    if len(dateParts) == 3 {
        day, _ := strconv.Atoi(dateParts[2])
        if day%2 == 1 {
            points += 6
        }
    }
    t, _ := time.Parse("15:04", receipt.PurchaseTime)
    if t.Hour() >= 14 && t.Hour() < 16 {
        points += 10
    }
    return points
}

// ProcessReceipt handles POST /receipts/process
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

// GetPoints handles GET /receipts/{id}/points
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

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Receipt Processor API is running!\n")
        fmt.Fprintln(w, "Usage Instructions:")
        fmt.Fprintln(w, "1. Submit a receipt:")
        fmt.Fprintln(w, "   curl -X POST http://localhost:8080/receipts/process -H \"Content-Type: application/json\" -d '{\"retailer\":\"Target\", \"purchaseDate\":\"2022-01-01\", \"purchaseTime\":\"13:01\", \"total\":\"35.35\", \"items\":[{\"shortDescription\":\"Mountain Dew 12PK\", \"price\":\"6.49\"}]}'")
        fmt.Fprintln(w, "2. Retrieve receipt points:")
        fmt.Fprintln(w, "   curl -X GET http://localhost:8080/receipts/{id}/points")
        fmt.Fprintln(w, "   (Replace {id} with the actual receipt ID from the previous response)")
    })
    
    http.HandleFunc("/receipts/process", ProcessReceipt)
    http.HandleFunc("/receipts/", GetPoints)
    fmt.Println("Server running on port 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

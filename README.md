# Receipt Processor
### Developed by Noah Caulfield (GitHub: [nwc6624](https://github.com/nwc6624))

This is a lightweight receipt processing web service written in Go. It allows users to submit receipts, assigns them a unique identifier, and calculates points based on a predefined set of rules.

## 🚀 Features

- **POST `/receipts/process`**: Accepts a receipt JSON and returns a unique receipt ID.
- **GET `/receipts/{id}/points`**: Retrieves the calculated points for a submitted receipt.

The service stores data **in-memory**, meaning that all receipts will be lost when the application restarts.

---

## 📌 Running the Service

### **Using Docker (Recommended)**
1. **Build the Docker Image:**
   ```sh
   docker build -t receipt-processor .
   ```

2. **Run the Container:**
   ```sh
   docker run -p 8080:8080 --name receipt-processor-container -d receipt-processor
   ```

3. **Check Running Containers (Optional):**
   ```sh
   docker ps
   ```

4. **View Logs (If Needed):**
   ```sh
   docker logs receipt-processor-container
   ```

---

### **Running Locally (Without Docker)**
#### Prerequisites:
- Go **1.18+**
- Ensure you have Go installed:  
  ```sh
  go version
  ```
#### Steps to Run:
1. Clone the repository:
   ```sh
   git clone https://github.com/nwc6624/receipt-processor-go.git
   cd receipt-processor-go
   ```

2. Build and start the server:
   ```sh
   go build -o receipt-processor
   ./receipt-processor
   ```

The service will be available at `http://localhost:8080`.

---

## 📡 API Documentation

### **📌 1. Submit a Receipt**
- **Endpoint:** `/receipts/process`
- **Method:** `POST`
- **Content-Type:** `application/json`
- **Request Body Example:**
  ```json
  {
    "retailer": "Target",
    "purchaseDate": "2022-01-01",
    "purchaseTime": "13:01",
    "total": "35.35",
    "items": [
      { "shortDescription": "Mountain Dew 12PK", "price": "6.49" },
      { "shortDescription": "Emils Cheese Pizza", "price": "12.25" }
    ]
  }
  ```
- **Response Example:**
  ```json
  { "id": "7fb1377b-b223-49d9-a31a-5a02701dd310" }
  ```

#### **💡 Example cURL Request**
```sh
curl -X POST http://localhost:8080/receipts/process      -H "Content-Type: application/json"      -d '{
         "retailer": "Target",
         "purchaseDate": "2022-01-01",
         "purchaseTime": "13:01",
         "total": "35.35",
         "items": [
             { "shortDescription": "Mountain Dew 12PK", "price": "6.49" },
             { "shortDescription": "Emils Cheese Pizza", "price": "12.25" }
         ]
     }'
```

---

### **📌 2. Retrieve Points for a Receipt**
- **Endpoint:** `/receipts/{id}/points`
- **Method:** `GET`
- **Response Example:**
  ```json
  { "points": 28 }
  ```

#### **💡 Example cURL Request**
```sh
curl -X GET http://localhost:8080/receipts/7fb1377b-b223-49d9-a31a-5a02701dd310/points
```

---

## 🎯 Points Calculation Rules

The points for each receipt are calculated based on the following rules:

1️⃣ **One point for every alphanumeric character in the retailer name.**  
2️⃣ **50 points if the total is a whole dollar amount (e.g., `$35.00`).**  
3️⃣ **25 points if the total is a multiple of `0.25`.**  
4️⃣ **5 points for every two items on the receipt.**  
5️⃣ **If an item description length (after trimming) is a multiple of `3`, multiply the price by `0.2`, round up, and add the result as points.**  
6️⃣ **5 bonus points if the total is greater than `$10.00` (for LLM-generated receipts).**  
7️⃣ **6 points if the purchase date is an **odd** day.**  
8️⃣ **10 points if the purchase time is between `2:00 PM - 4:00 PM`.**  

---

## 🧪 Testing Examples

### ✅ **Test 1: Walgreens Receipt**
#### **Request**
```sh
curl -X POST http://localhost:8080/receipts/process      -H "Content-Type: application/json"      -d '{
         "retailer": "Walgreens",
         "purchaseDate": "2022-01-02",
         "purchaseTime": "08:13",
         "total": "2.65",
         "items": [
             {"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
             {"shortDescription": "Dasani", "price": "1.40"}
         ]
     }'
```
#### **Expected Points Calculation**
| Rule | Points Earned |
|-----------------|--------------|
| `Walgreens` has **9** alphanumeric characters | **9** |
| `2.65` is **not** a whole dollar | **0** |
| `2.65` is **not** a multiple of 0.25 | **0** |
| **2 items** → `(2 / 2) * 5 = 5` | **5** |
| `Dasani` (6 chars) qualifies for price multiplier → `ceil(1.40 * 0.2) = 1` | **1** |
| Purchase day is **even (2)** | **0** |
| Purchase time **not between 2-4 PM** | **0** |
| **Total Points** | **15** |

---

### ✅ **Test 2: Target Receipt**
#### **Request**
```sh
curl -X POST http://localhost:8080/receipts/process      -H "Content-Type: application/json"      -d '{
         "retailer": "Target",
         "purchaseDate": "2022-01-02",
         "purchaseTime": "13:13",
         "total": "1.25",
         "items": [
             {"shortDescription": "Pepsi - 12-oz", "price": "1.25"}
         ]
     }'
```
#### **Expected Points Calculation**
| Rule | Points Earned |
|-----------------|--------------|
| `Target` has **6** alphanumeric characters | **6** |
| `1.25` is **not** a whole dollar | **0** |
| `1.25` **is** a multiple of 0.25 | **25** |
| **1 item** → `(1 / 2) * 5 = 0` | **0** |
| `Pepsi - 12-oz` (14 chars) **not** a multiple of 3 | **0** |
| Purchase day is **even (2)** | **0** |
| Purchase time **not between 2-4 PM** | **0** |
| **Total Points** | **31** |

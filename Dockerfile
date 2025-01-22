# Use official Golang image
FROM golang:1.19

# Set the working directory
WORKDIR /app

# Copy source code
COPY . .

# Build the application
RUN go mod init receipt-processor && go mod tidy && go build -o receipt-processor

# Expose port
EXPOSE 8080

# Run the application
CMD ["./receipt-processor"]

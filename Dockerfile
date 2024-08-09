# Dockerfile for Calendar service

# Use a Golang base image
FROM golang:1.20-alpine

# Set the working directory
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o calendar cmd/calendar/main.go

# Expose the port the service runs on
EXPOSE 8080

# Command to run the application
CMD ["./calendar"]

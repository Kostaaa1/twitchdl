# Start with the official Golang image
FROM golang:1.17-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the Go app
RUN go build -o twitch cmd/twitch/main.go

# Start a new stage from scratch
FROM alpine:latest

# Install Node.js and npm (if needed)
RUN apk add --no-cache nodejs npm

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go executable
COPY --from=build /app/twitch /app/twitch

# Copy the scripts directory (assuming you need it)
COPY scripts /app/scripts

# Set environment variables if necessary
# ENV KEY=VALUE

# Command to run your application
CMD ["./twitch"]

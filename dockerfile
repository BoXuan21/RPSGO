FROM golang:1.21-alpine

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main .

# Run the application
CMD ["./main"]
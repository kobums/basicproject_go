FROM --platform=linux/amd64 golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copying .env.yml to the CURRENT working directory (/root/)
COPY .env.yml .env.yml

# Creating webdata directory in the CURRENT working directory
RUN mkdir -p webdata


ENV APP_MODE=production

CMD ["./main"]
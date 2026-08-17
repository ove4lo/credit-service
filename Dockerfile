# Multi-stage build
FROM golang:1.25 AS builder
WORKDIR /app # work folder in the container

# Before the rest of the code is copied, dependencies are downloaded as a separate layer and cached.
# Docker won't re-download them during a rebuild unless `go.mod` changes
COPY go.mod go.sum ./
RUN go mod download

# Copy
COPY . .

# Build a "clean" binary with no dependencies on system libraries
RUN CGO_ENABLED=0 go build -o /creditservice ./cmd/creditservice

# Final image
FROM alpine:latest
COPY --from=builder /creditservice /creditservice
EXPOSE 4000
CMD ["/creditservice"]
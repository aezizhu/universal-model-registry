# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Run stage
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server
ENV MCP_TRANSPORT=sse
ENV PORT=8000
EXPOSE 8000
ENTRYPOINT ["/server"]

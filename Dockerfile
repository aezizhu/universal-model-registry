# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go-server/go.mod go-server/go.sum ./
RUN go mod download
COPY go-server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Run stage
FROM alpine:3.20
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /server /server
USER appuser
ENV MCP_TRANSPORT=sse
ENV PORT=8000
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:${PORT}/sse || exit 1
ENTRYPOINT ["/server"]

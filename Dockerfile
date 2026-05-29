# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./dist/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o hltv-mcp github.com/arcdent/hltv-mcp

# Stage 3: Runtime
# chromedp/headless-shell provides a headless Chrome instance for chromedp
FROM chromedp/headless-shell:latest
WORKDIR /
# ca-certificates is required for Go's net/http to verify TLS connections
# (e.g., LLM translate proxy). chromedp/headless-shell includes openssl
# but not the CA certificate bundle.
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /data
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/headless-shell/headless-shell
ENV FIRECRAWL_API_KEY=
VOLUME ["/data"]
ENTRYPOINT ["/hltv-mcp"]

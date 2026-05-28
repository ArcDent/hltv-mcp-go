# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto
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
RUN mkdir -p /data
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/headless-shell/headless-shell
VOLUME ["/data"]
ENTRYPOINT ["/hltv-mcp"]

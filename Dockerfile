# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/frontend/dist ./dist/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o hltv-mcp github.com/arcdent/hltv-mcp

# Stage 3: Runtime
# chrome-headless-shell:stable provides chrome-headless-shell binary
# Go starts it via chromedp.ExecAllocator at runtime
FROM chrome-headless-shell:stable
COPY --from=builder /app/hltv-mcp /hltv-mcp
EXPOSE 8082
ENV HTTP_PORT=8082
ENV HTTP_HOST=0.0.0.0
ENV HLTV_CHROME_PATH=/usr/bin/chrome-headless-shell
ENTRYPOINT ["/hltv-mcp"]

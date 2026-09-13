# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------
# Stage 1: Go Builder (multi-stage binary compilation)
# ---------------------------------------------------------------------
FROM golang:1.24-alpine AS go-builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./services/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/worker ./services/worker

# ---------------------------------------------------------------------
# Target: API Service
# ---------------------------------------------------------------------
FROM alpine:3.20 AS api
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /bin/api /app/api

EXPOSE 8080
ENTRYPOINT ["/app/api"]

# ---------------------------------------------------------------------
# Target: Worker Service
# ---------------------------------------------------------------------
FROM alpine:3.20 AS worker
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /bin/worker /app/worker

EXPOSE 8081
ENTRYPOINT ["/app/worker"]

# ---------------------------------------------------------------------
# Target: Frontend Builder
# ---------------------------------------------------------------------
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---------------------------------------------------------------------
# Target: Frontend Nginx Server
# ---------------------------------------------------------------------
FROM nginx:alpine AS frontend
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

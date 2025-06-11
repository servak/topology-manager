FROM node:18-alpine AS frontend-builder

WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o topology-manager ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

COPY --from=backend-builder /app/topology-manager .
COPY --from=frontend-builder /app/web/build ./web/build
COPY config/ ./config/

ENV WEB_DIR=/app/web/build

EXPOSE 8080

CMD ["./topology-manager", "api"]
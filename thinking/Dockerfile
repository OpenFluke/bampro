# 1. Build Go backend
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the Go app
COPY . .

RUN go build -o main .

# 2. Build React dashboard
FROM node:20-alpine AS dashboard-builder

WORKDIR /dashboard

# Only copy package files first for Docker cache
COPY bampro-dashboard/package.json bampro-dashboard/package-lock.json ./
RUN npm install

# Copy the rest of the dashboard source
COPY bampro-dashboard ./
RUN npm run build

# 3. Final container (serves Go + static + dashboard)
FROM alpine:latest

WORKDIR /app

# Copy Go backend binary
COPY --from=builder /app/main .

# Copy dashboard build output to static folder
COPY --from=dashboard-builder /dashboard/build ./static/dashboard

# Copy backend's static files (CSS, images, etc.)
COPY static ./static

EXPOSE 8123
EXPOSE 9001


CMD ["./main"]

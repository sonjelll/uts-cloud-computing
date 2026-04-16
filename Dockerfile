# Stage 1: Build
FROM golang:alpine AS builder
WORKDIR /app

# Copy module files
COPY go.mod ./
# Ignore error if go.sum doesn't exist yet
RUN go mod download || true

# Copy seluruh source code dan folder templates
COPY . .

# Build aplikasi Golang
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-computing-app

# Stage 2: Run
FROM alpine:latest
WORKDIR /app

# Copy file binary dari tahap builder
COPY --from=builder /cloud-computing-app /app/cloud-computing-app

# COPY FOLDER TEMPLATES (Ini yang bikin error tadi!)
COPY --from=builder /app/templates /app/templates

ENV PORT=8080
EXPOSE 8080

CMD ["/app/cloud-computing-app"]
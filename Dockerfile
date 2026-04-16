# Stage 1: Build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod init cloud-computing-app
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-computing-app

# Stage 2: Run
FROM alpine:latest

WORKDIR /app
COPY --from=builder /cloud-computing-app /app/cloud-computing-app

ENV PORT=8080
ENV NAME="Cloud Computing Students"

EXPOSE 8080

CMD ["/app/cloud-computing-app"]
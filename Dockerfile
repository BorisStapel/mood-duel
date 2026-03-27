# ---- Build stage ----
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mood-duel .

# ---- Run stage ----
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/mood-duel .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./mood-duel"]

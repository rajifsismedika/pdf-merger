FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o bin/pdfmerge-server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/bin/pdfmerge-server .
COPY --from=builder /app/.env .

EXPOSE 8585

CMD ["./pdfmerge-server", "-port", "8585"]

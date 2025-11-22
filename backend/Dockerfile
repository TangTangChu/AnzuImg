FROM library/golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o anzuimg ./backend/cmd/server


FROM alpine:3.20

WORKDIR /app

RUN mkdir -p /data/images

COPY --from=builder /app/anzuimg /app/anzuimg

EXPOSE 8080

CMD ["/app/anzuimg"]

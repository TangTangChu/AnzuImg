FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN echo "https://mirrors.ustc.edu.cn/alpine/edge/main" > /etc/apk/repositories && \
    echo "https://mirrors.ustc.edu.cn/alpine/edge/community" >> /etc/apk/repositories && \
    echo "https://mirrors.ustc.edu.cn/alpine/edge/testing" >> /etc/apk/repositories

RUN apk update && apk add --no-cache \
    vips-dev \
    libwebp-dev \
    build-base \
    pkgconf

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o anzuimg ./backend/cmd/server

FROM alpine:edge

WORKDIR /app

RUN echo "https://mirrors.ustc.edu.cn/alpine/edge/main" > /etc/apk/repositories && \
    echo "https://mirrors.ustc.edu.cn/alpine/edge/community" >> /etc/apk/repositories && \
    echo "https://mirrors.ustc.edu.cn/alpine/edge/testing" >> /etc/apk/repositories

RUN apk update && apk add --no-cache \
    vips \
    vips-tools \
    libheif \
    libwebp \
    ca-certificates \
    tzdata \
    libjxl \
    && mkdir -p /data/images

COPY --from=builder /app/anzuimg /app/anzuimg

EXPOSE 8080

CMD ["/app/anzuimg"]
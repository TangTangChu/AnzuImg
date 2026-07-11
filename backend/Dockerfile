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
    ffmpeg \
    ca-certificates \
    tzdata \
    libjxl \
    && addgroup -S -g 10001 anzuimg \
    && adduser -S -D -H -u 10001 -G anzuimg anzuimg \
    && mkdir -p /data/images /data/logs /data/tmp \
    && chown -R anzuimg:anzuimg /app /data

COPY --from=builder --chown=anzuimg:anzuimg /app/anzuimg /app/anzuimg

USER anzuimg

EXPOSE 8080

CMD ["/app/anzuimg"]

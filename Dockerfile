FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN curl -sL https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64 -o /usr/local/bin/tailwindcss     && chmod +x /usr/local/bin/tailwindcss     && echo '@tailwind base; @tailwind components; @tailwind utilities;' > /tmp/tw-input.css     && tailwindcss -i /tmp/tw-input.css -o web/tailwind.css --content 'web/index.html' --minify

RUN CGO_ENABLED=0 go build -o control ./cmd/control
RUN CGO_ENABLED=0 go build -o agent ./cmd/agent

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache wireguard-tools iproute2 openssl

COPY --from=builder /app/control .
COPY --from=builder /app/agent .
COPY --from=builder /app/web ./web

RUN adduser --disabled-password --gecos "" appuser
USER appuser

EXPOSE 8080
EXPOSE 51820/udp

CMD ["./control"]

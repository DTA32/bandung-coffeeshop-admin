FROM golang:1.26 AS builder
ARG TARGETARCH
ARG TEMPL_VERSION=v0.3.1020
ARG TAILWIND_VERSION=v4.3.3
WORKDIR /app

# Tailwind v4 standalone CLI (no Node), pinned and checksum-verified.
# Retries cover flaky runner -> GitHub release downloads (connection resets).
RUN set -eu; \
    case "$TARGETARCH" in \
      amd64) asset=tailwindcss-linux-x64;   sum=dc61b3ac6b8c9ca874c0cc4c57b2409791a64c5540404ca5f5367360babc313a ;; \
      arm64) asset=tailwindcss-linux-arm64; sum=55fd0b241214eff3de1e8ee4f22796662f2d2e7a49bcfca7477cfd0bac398195 ;; \
      *) echo "unsupported arch $TARGETARCH"; exit 1 ;; \
    esac; \
    curl -sSfL --retry 5 --retry-all-errors --retry-delay 5 --connect-timeout 20 --max-time 120 \
      -o /usr/local/bin/tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/${asset}"; \
    echo "${sum}  /usr/local/bin/tailwindcss" | sha256sum -c -; \
    chmod +x /usr/local/bin/tailwindcss

COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go install github.com/a-h/templ/cmd/templ@${TEMPL_VERSION}

COPY . .
RUN templ generate \
 && tailwindcss -i static/css/input.css -o static/css/app.css --minify \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /app/main .

FROM alpine:3.22
WORKDIR /app
RUN adduser -D -H -u 10001 app
COPY --from=builder /app/main ./
USER app

ARG APP_PORT=8080
ENV APP_PORT=${APP_PORT}
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:${APP_PORT:-8080}/health || exit 1

CMD ["/app/main"]

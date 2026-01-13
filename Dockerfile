# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.24
ARG ALPINE_VERSION=3.20

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS debug
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN go install github.com/go-delve/delve/cmd/dlv@v1.23.1

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -gcflags="all=-N -l" -o /out/server ./cmd/server

EXPOSE 8080 40000
CMD ["/go/bin/dlv", "exec", "/out/server", "--headless", "--listen=:40000", "--api-version=2", "--accept-multiclient", "--log"]

FROM alpine:${ALPINE_VERSION} AS runtime
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
  && addgroup -S app && adduser -S -G app app

COPY --from=build /out/server /app/server
COPY web /app/web

USER app
EXPOSE 8080
CMD ["/app/server"]
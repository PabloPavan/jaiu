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
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/renewal-worker ./cmd/renewal-worker

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS dev
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata
ENV PATH="/go/bin:${PATH}"

RUN go install github.com/air-verse/air@v1.52.3
RUN go install github.com/a-h/templ/cmd/templ@v0.3.977
RUN go install github.com/go-delve/delve/cmd/dlv@v1.26.0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

EXPOSE 8080 40000
CMD ["air", "-c", ".air.toml"]

FROM alpine:${ALPINE_VERSION} AS runtime
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
  && addgroup -S app && adduser -S -G app app

COPY --from=build /out/server /app/server
COPY --from=build /out/renewal-worker /app/renewal-worker
COPY web /app/web

USER app
EXPOSE 8080
CMD ["/app/server"]

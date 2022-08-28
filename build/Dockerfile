# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS builder

WORKDIR /code

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build
ENV APP_PORT 3000
ARG PRODUCT_NAME
ARG PRODUCT_VER

COPY go.mod go.sum ./cmd/$PRODUCT_NAME/ ./internal/ ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/cache \
    --mount=type=cache,target=/go-build \
    go build -ldflags="-X 'main.Version=$PRODUCT_VER'" -o bin/$PRODUCT_NAME .

CMD ["/code/bin/$PRODUCT_NAME"]

#FROM scratch
FROM alpine
ENV APP_PORT 3000
ARG PRODUCT_NAME

RUN apk --no-cache add curl
COPY --from=builder /code/bin/$PRODUCT_NAME /usr/local/bin/go-metr
WORKDIR /usr/local/bin/
HEALTHCHECK CMD curl --fail http://localhost:$APP_PORT/status || exit 1
CMD ["/usr/local/bin/go-metr"]
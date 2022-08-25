# syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS builder

WORKDIR /code

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build
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

FROM scratch
ARG PRODUCT_NAME
COPY --from=builder /code/bin/$PRODUCT_NAME /usr/local/bin/go-metr
WORKDIR /usr/local/bin/
CMD ["/usr/local/bin/go-metr"]
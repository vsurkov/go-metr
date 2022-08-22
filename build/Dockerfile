## syntax=docker/dockerfile:1.4
FROM --platform=$BUILDPLATFORM golang:1.18-alpine AS builder

WORKDIR /code

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build
ARG PRODUCT_NAME

COPY go.mod go.sum ./cmd/$PRODUCT_NAME/ ./internal/ ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/cache \
    --mount=type=cache,target=/go-build \
    go build -o bin/$PRODUCT_NAME .

CMD ["/code/bin/$PRODUCT_NAME"]

FROM builder as dev-envs

RUN <<EOF \
apk update \
apk add git \
EOF

RUN <<EOF \
addgroup -S docker \
adduser -S --shell /bin/bash --ingroup docker vscode \
EOF

# install Docker tools (cli, buildx, compose)
COPY --from=gloursdocker/docker / /

CMD ["go", "run", "main.go"]

FROM scratch
COPY --from=builder /code/bin/$PRODUCT_NAME /usr/local/bin/$PRODUCT_NAME
CMD ["/usr/local/bin/$PRODUCT_NAME"]
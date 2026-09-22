# syntax=docker/dockerfile:1

FROM golang:1.27.1-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY example/go.mod example/go.sum ./example/

RUN --mount=type=cache,target=/go/pkg/mod \
    cd example && go mod download

COPY pkg ./pkg
COPY example ./example

WORKDIR /src/example

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w -buildid=" \
    -o /out/gor-example \
    .

FROM scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/gor-example /app/gor-example
COPY --from=build /src/example/static /app/static

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/gor-example"]

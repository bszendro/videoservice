# syntax=docker/dockerfile:1

FROM golang:1.26 AS build
WORKDIR /src

# Download deps first so this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/videoservice .

# Static binary, no shell/package manager, runs as non-root (uid 65532).
FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source=https://github.com/bszendro/videoservice

COPY --from=build /out/videoservice /videoservice
EXPOSE 8000
EXPOSE 7000
USER nonroot:nonroot
ENTRYPOINT ["/videoservice"]

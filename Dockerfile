FROM golang:1.26.4 AS builder

ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -o /out/groutdra ./cmd/groutdra

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/groutdra /usr/bin/groutdra
ENTRYPOINT ["/usr/bin/groutdra"]

FROM golang:1@sha256:c7c94588b6445f5254fbc34df941afa10de04706deb330e62831740c9f0f2030 AS builder

WORKDIR /workspace/src

ENV CGO_ENABLED=0 \
    GOOS=${TARGET_OS} \
    GOARCH=${TARGET_ARCH}

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,target=. \
    go build -o ../app main.go

FROM gcr.io/distroless/static:nonroot@sha256:80c956fb0836a17a565c43a4026c9c80b2013c83bea09f74fa4da195a59b7a99

WORKDIR /
COPY --from=builder /workspace/app .
USER 65532:65532
ENTRYPOINT ["/app"]

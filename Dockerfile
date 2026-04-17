FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git make gawk && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /build

# Copy source first, then generate go.mod/go.sum inside container
COPY . .

# Initialize go.mod if not present, then tidy to generate go.sum
RUN if [ ! -f go.mod ]; then \
    go mod init github.com/glemsom/dkvmmanager; \
    fi && \
    go mod tidy

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 go build \
    -ldflags "-X github.com/glemsom/dkvmmanager/internal/version.Version=${VERSION} \
              -X github.com/glemsom/dkvmmanager/internal/version.Commit=${COMMIT} \
              -X github.com/glemsom/dkvmmanager/internal/version.Date=${DATE} \
              -extldflags '-static'" \
    -o dkvmmanager .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates gawk go yq jq

WORKDIR /app

COPY --from=builder /build/dkvmmanager /usr/local/bin/
COPY --from=builder /go/bin/golangci-lint /usr/local/bin/

ENTRYPOINT ["dkvmmanager"]

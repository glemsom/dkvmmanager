FROM --platform=$BUILDPLATFORM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS builder

# Pin tool versions for reproducible builds
ARG GOLANGCI_LINT_VERSION=v2.2.0

RUN apk add --no-cache git make gawk && \
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

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

FROM --platform=$BUILDPLATFORM alpine:3.19 AS runtime

RUN apk add --no-cache ca-certificates gawk go yq jq

WORKDIR /app

COPY --from=builder /build/dkvmmanager /usr/local/bin/
COPY --from=builder /go/bin/golangci-lint /usr/local/bin/

ENTRYPOINT ["dkvmmanager"]

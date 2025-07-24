ARG GO_VERSION=1.24-alpine
ARG ALPINE_VERSION=latest

# Build stage
FROM docker.io/library/golang:${GO_VERSION} AS builder

WORKDIR /build

COPY src/go.mod src/go.sum .
RUN go mod download

COPY src/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o patternizer .

# Runtime stage
FROM docker.io/library/alpine:${ALPINE_VERSION}

RUN apk --no-cache add git

COPY --from=builder /build/patternizer /usr/local/bin/patternizer

ARG PATTERNIZER_RESOURCES_DIR=/tmp/resources
WORKDIR ${PATTERNIZER_RESOURCES_DIR}

COPY resources/* .

ARG PATTERN_REPO_ROOT=/repo
WORKDIR ${PATTERN_REPO_ROOT}

ENV PATTERNIZER_RESOURCES_DIR=${PATTERNIZER_RESOURCES_DIR}

ENTRYPOINT ["patternizer"]
CMD ["help"]

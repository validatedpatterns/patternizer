ARG GO_VERSION=1.26.3
ARG GOARCH=amd64

# Build stage
FROM registry.access.redhat.com/ubi10/go-toolset:${GO_VERSION} AS builder

WORKDIR /build

COPY src/go.mod src/go.sum .
RUN go mod download

COPY src/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -a -installsuffix cgo -o patternizer .

# Runtime stage
FROM registry.access.redhat.com/ubi10/ubi-minimal:10.0

COPY --from=builder /build/patternizer /usr/local/bin/patternizer

ENTRYPOINT ["patternizer"]
CMD ["help"]

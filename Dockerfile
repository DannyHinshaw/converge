# ARGs provide default versions for base images and tooling used in this multi-stage build.
# These values are primarily defined in the root .env file for consistency across builds.
# If no value is passed during build, these defaults will be used.
ARG ALPINE_IMAGE
ARG GOLANG_IMAGE
ARG GOFUMPT_VERSION
ARG GOLANGCI_LINT_VERSION
ARG PKGSITE_VERSION

# Create pinned alpine base image
FROM ${ALPINE_IMAGE} AS alpine-builder

# Create pinned golang base image
FROM ${GOLANG_IMAGE} AS golang-builder
RUN apk add --no-cache git gcc musl-dev

# gofumpt stage: used for stricter Go code formatting, an extension of gofmt.
FROM golang-builder AS gofumpt
ARG GOFUMPT_VERSION
RUN go install mvdan.cc/gofumpt@$GOFUMPT_VERSION

# golangci-lint stage: used for linting Go code, pulled from the official repository.
FROM alpine-builder AS golangci
ARG GOLANGCI_LINT_VERSION
# Download and install golangci-lint using the official install script (recommended).
RUN wget -nv -O - https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s $GOLANGCI_LINT_VERSION

# pkgsite stage: used to generate documentation for Go packages.
FROM golang-builder AS pkgsite
ARG PKGSITE_VERSION
RUN go install golang.org/x/pkgsite/cmd/pkgsite@$PKGSITE_VERSION

# tools target stage contains all tool binaries from the preceeding build stages.
FROM golang-builder as tools
COPY --from=gofumpt /go/bin/gofumpt /usr/bin
COPY --from=golangci /bin/golangci-lint /usr/bin
COPY --from=pkgsite /go/bin/pkgsite /usr/bin
WORKDIR /workspace

# When `docker compose` commands are ran,
# these "_VERSION" ARGS are populated by
# the .env file in the root directory, and
# made available at image build time by the
# service.build.args fields set in the
# compose.yaml file.
ARG ALPINE_IMAGE_VERSION
ARG GOLANG_IMAGE_VERSION
ARG GOFUMPT_VERSION
ARG GOLANGCI_LINT_VERSION
ARG PKGSITE_VERSION

# Create pinned alpine base image
ARG ALPINE_IMAGE_VERSION
FROM $ALPINE_IMAGE_VERSION as alpine-builder

# Create pinned golang base image
ARG GOLANG_IMAGE_VERSION
FROM $GOLANG_IMAGE_VERSION as golang-builder

# gofumpt used for stricter gofmt code formatting.
FROM alpine-builder as gofumpt
ARG GOFUMPT_VERSION
RUN wget -nv -O /bin/gofumpt \
    https://github.com/mvdan/gofumpt/releases/download/$GOFUMPT_VERSION/gofumpt_${GOFUMPT_VERSION}_linux_arm64 \
    && chmod +x /bin/gofumpt

# golangci-lint is used for Go code linting.
FROM alpine-builder as golangci
ARG GOLANGCI_LINT_VERSION
RUN wget -nv -O - https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
    | sh -s $GOLANGCI_LINT_VERSION

# pkgsite is used for generating the pkg.go.dev site.
FROM golang-builder as pkgsite
ARG PKGSITE_VERSION
RUN go install golang.org/x/pkgsite/cmd/pkgsite@$PKGSITE_VERSION

# tools target stage contains all tool binaries from the preceeding build stages.
FROM golang-builder as tools

RUN apk add --no-cache gcc musl-dev

COPY --from=gofumpt /bin/gofumpt /usr/bin
COPY --from=golangci /bin/golangci-lint /usr/bin
COPY --from=pkgsite /go/bin/pkgsite /usr/bin

WORKDIR /workspace

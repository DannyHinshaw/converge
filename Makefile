# /////////////////////////////////////////////////////////////
#                            Makefile
# \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

# Set the default command of the Makefile to be
# the "help" command (printing the command docs).
# So the following commands will both print the docs:
# `make` and `make help`.
.DEFAULT_GOAL := help

# Set the name of the app
APP_NAME := converge

# COMPOSE contains the base Docker Compose command for running various
# binaries in the tools Compose service as ephemeral containers.
COMPOSE = docker compose

# TOOLS contains the base Docker Compose command for running various
# tools in the tools Compose service as ephemeral containers.
# It can be overridden by setting the TOOLS environment variable, for example:
# `make lint TOOLS=` will run the linter locally instead of in a container.
# However, this is not recommended as it will not be consistent with CI.
#
# Additionally, if you prefer to run tools locally, you will need to install
# (and manage) them on your machine. There are simply too many variables
# at play with developer machines to ensure consistency across all engineering.
TOOLS ?= $(COMPOSE) run --rm --service-ports tools

# GOFUMPT contains the base Go command for running gofumpt
# defaulting to running it in the tools container.
GOFUMPT ?= $(TOOLS) gofumpt

# GOLINT contains the base Go command for running golangci-lint
# defaulting to running it in the tools container.
GOLINT ?= $(TOOLS) golangci-lint

# PKGSITE contains the base Go command for running pkgsite,
# a browser-based tool for viewing Go documentation.
PKGSITE := $(TOOLS) pkgsite

# GO contains the base Go command for running go
# defaulting to running it in the tools container.
# This can be overridden by either setting the GO
# environment variable or by setting the TOOLS
# environment variable to an empty string.
GO ?= $(TOOLS) go

# GOTEST contains the base Go command for running tests.
# It can be overridden by setting the GOTEST environment variable.
GOTEST ?= $(GO) test

# GOVET contains the base Go command for running go vet.
# It can be overridden by setting the GOVET environment variable.
GOVET ?= $(GO) vet

# Set and export the $GOPATH env var if not already set.
export GOPATH ?= $(shell go env GOPATH)

###################
#   Main Targets  #
###################

.PHONY: all
## runs all the things
all: tidy fmt/fix verify build/cli test/full

.PHONY: deps
## checks to make sure all general are present on the machine for use in other make targets
deps:
	@MISSING=""; \
	for cmd in docker go; do \
	  command -v $$cmd >/dev/null 2>&1 || MISSING="$$MISSING $$cmd"; \
	done; \
	if [ "$$MISSING" != "" ]; then \
	  echo "Make: the following dependencies are required but couldn't be found:"; \
	  echo "$$MISSING"; \
	  echo " Please install them and re-run 'make deps'"; \
	  exit 1; \
  	else \
  	  go mod download; \
  	fi

.PHONY: build
## builds all binaries/images
build: build/cli compose/build

.PHONY: build/cli
## builds the converge CLI binary
build/cli:
	@mkdir -p bin
	@VERSION=$$(git describe --tags --always || echo "(dev)") \
		&& echo "building $(APP_NAME) $$VERSION" \
		&& go build -v -trimpath -ldflags "-X main.version=$$VERSION" -o bin/$(APP_NAME)

.PHONY: compose/build
## builds resources
compose/build: deps
	@$(COMPOSE) build --no-cache

.PHONY: compose/clean
## cleans up resources
compose/clean:
	@$(COMPOSE) down --rmi=all --remove-orphans --volumes

########################
#    Linting/Verify    #
########################

.PHONY: verify
## runs all code verification tools
verify: lint vet

.PHONY: lint
## runs all code linters
lint:
	@$(GOLINT) run

.PHONY: vet
## runs go vet on all source files
vet:
	@$(GOVET) ./...

################
#    Format    #
################

.PHONY: fmt/check
## checks code formatting on all source files and errors if bad formatting is detected
fmt/check:
	@$(GOFUMPT) -extra -d .

.PHONY: fmt/fix
## runs gofumpt code formatter on all source files and fixes any formatting issues
fmt/fix:
	@$(GOFUMPT) -extra -l -w .

################
#     Test     #
################

# Default to running tests with the -race flag enabled,
# but allow the user to disable it by setting the RACE
# environment variable to false. e.g.: `RACE=false make test`
RACE ?= true
RACE_FLAG := $(if $(filter true,$(RACE)),-race,)
TEST_FLAGS := -v -timeout 1m
COVER_FLAGS :=

.PHONY: test
## runs all the Go unit tests in the repo
test:
	@$(GOTEST) $(RACE_FLAG) $(TEST_FLAGS) $(COVER_FLAGS) ./...

.PHONY: test/cover
## runs all the tests with coverage enabled
test/cover: COVER_FLAGS = -coverprofile=coverage.out -covermode=atomic
test/cover: test

.PHONY: test/full
## runs all the tests with coverage enabled and opens the coverage report in the browser
test/full: test/cover
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@open coverage.html


################
#   Release    #
################

.PHONY: tidy
## runs go mod tidy
tidy:
	@$(GO) mod tidy

.PHONY: release
## Issues a new release with git tag. Example usage: make release VERSION=v1.0.0
release:
	# Ensure a version is provided
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set. Use make release VERSION=vx.y.z"; \
		exit 1; \
	fi
	git tag $(VERSION)
	git push origin $(VERSION)


################
#     Halp     #
################

.PHONY: docs
## starts the pkgsite server and opens the browser to the pkgsite page
docs:
	@echo "once server is running, visit the following url in your browser:"
	@echo "http://localhost:3030"
	@$(PKGSITE) -http=0.0.0.0:3030

.PHONY: help
## prints out the help documentation (also will be printed by simply running `make` command with no arg)
help: Makefile
	@# avert your eyes...
	@echo "$$(tput bold)Available commands:$$(tput sgr0)";echo;sed -ne"/^## /{h;s/.*//;:d" -e"H;n;s/^## //;td" -e"s/:.*//;G;s/\\n## /---/;s/\\n/ /g;p;}" ${MAKEFILE_LIST}|awk -F --- -v n=$$(tput cols) -v i=25 -v a="$$(tput setaf 6)" -v z="$$(tput sgr0)" '{printf"%s%*s%s ",a,-i,$$1,z;m=split($$2,w," ");l=n-i;for(j=1;j<=m;j++){l-=length(w[j])+1;if(l<= 0){l=n-i-length(w[j])-1;printf"\n%*s ",-i," ";}printf"%s ",w[j];}printf"\n";}'|more $(shell test $(shell uname) == Darwin && echo '-Xr')


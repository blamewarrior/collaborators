NAME := collaborators
PREFIX ?= usr/local
VERSION := $$(git rev-parse HEAD | cut -c -6)
GOVERSION := $(shell go version)
BUILDDATE := $(shell date -u +"%B %d, %Y")
BUILDER := $(shell echo "`git config user.name` <`git config user.email`>")
PKG_RELEASE ?= 1
PROJECT_URL := "https://github.com/andrewslotin/$(NAME)"
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.buildDate=$(BUILDDATE)' \
           -X 'main.builder=$(BUILDER)' \
           -X 'main.buildGoVersion=$(GOVERSION)'

# development tasks
PACKAGES := $$(go list ./... | grep -v /vendor/ | grep -v /cmd/)
test: setupdb
	@echo "Running tests..."
	DB_USER=postgres DB_NAME=bw_collaborators_test go test $(PACKAGES)

benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. $(PACKAGES)

setupdb:
	@echo "Setting up test database..."
	psql -U postgres -c "DROP DATABASE IF EXISTS bw_collaborators_test;"
	psql -U postgres -c "CREATE DATABASE bw_collaborators_test;"
	psql -U postgres bw_collaborators_test <db/schema.sql

# build tasks
SOURCES := $(shell find . -type f \( -name '*.go' -and -not -name '*_test.go' \))
build: $(SOURCES)
	go build -ldflags "$(LDFLAGS)" -o $(NAME)

all: test build
.DEFAULT_GOAL := all

clean:
	go clean

.PHONY: all build test benchmark clean setupdb

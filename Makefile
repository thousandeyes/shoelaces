GO = go
SCDOC = scdoc
VERSION ?= dev
LDFLAGS = "-s -w -X main.version=$(VERSION)"

pkgs = $(shell $(GO) list ./... | grep -v /vendor/)

all:
	$(GO) build

run:
	$(GO) run . -data-dir test/integ-test/integ-test-configs -debug

fmt:
	$(GO) fmt

clean:
	rm -f shoelaces docs/shoelaces.8

shoelaces.8:
	$(SCDOC) < docs/shoelaces.8.scd > docs/shoelaces.8

docs: shoelaces.8

test: fmt
	$(GO) test -v $(pkgs) && \
		./test/integ-test/integ_test.py -vv

.PHONY: all run clean docs

binaries: linux windows
linux:
	SOURCE_DATE_EPOCH=$(shell git log -1 --format=%ct) \
	CGO_ENABLED=0 GOOS=linux ${GO} build -trimpath -o bin/shoelaces -ldflags ${LDFLAGS}
windows:
	CGO_ENABLED=0 GOOS=windows ${GO} build -trimpath -o bin/shoelaces.exe -ldflags ${LDFLAGS}

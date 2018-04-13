GO = go
pkgs = $(shell $(GO) list ./... | grep -v /vendor/)

all:
	$(GO) build

fmt:
	$(GO) fmt

clean:
	rm -f shoelaces

test: fmt
		$(GO) test -v $(pkgs) && \
			./test/integ-test/integ_test.py

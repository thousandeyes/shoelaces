GO = go
SCDOC = scdoc
LDFLAGS = "-s -w"

pkgs = $(shell $(GO) list ./... | grep -v /vendor/)

all:
	$(GO) build

fmt:
	$(GO) fmt

clean:
	rm -f shoelaces docs/shoelaces.8

shoelaces.8:
	$(SCDOC) < docs/shoelaces.8.scd > docs/shoelaces.8

docs: shoelaces.8

test: fmt
		$(GO) test -v $(pkgs) && \
			./test/integ-test/integ_test.py
# Internal information, in case you are hinted to use `-vv`
# $ tail -n 2 test/integ-test/integ_test.py
# if __name__ == "__main__":
#     pytest.main(args=['-v'], plugins=None)

.PHONY: all clean docs

binaries: linux windows
linux:
		GOOS=linux ${GO} build -o bin/shoelaces -ldflags ${LDFLAGS}
windows:
		GOOS=windows ${GO} build -o bin/shoelaces.exe -ldflags ${LDFLAGS}

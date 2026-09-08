BINARY  := tagger
PREFIX  ?= /usr/local
BINDIR  := $(PREFIX)/bin

.PHONY: all build test vet fmt clean install

all: build

build:
	go build -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)

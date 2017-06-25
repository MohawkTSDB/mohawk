PREFIX := $(GOPATH)
BINDIR := $(PREFIX)/bin
SOURCE := *.go router/*.go middleware/*.go middleware/gzip/*.go backend/*.go backend/example/*.go backend/memory/*.go backend/sqlite/*.go

all: mohawk

mohawk: $(SOURCE)
	go build -o mohawk *.go

.PHONY: fmt
fmt: $(SOURCE)
	gofmt -s -l -w $(SOURCE)

.PHONY: clean
clean:
	$(RM) mohawk

.PHONY: install
install:
	install -D -m0755 mohawk $(DESTDIR)/$(BINDIR)/mohawk

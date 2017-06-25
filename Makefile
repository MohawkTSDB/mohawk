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

.PHONY: test
test:
	bats test/mohawk.bats

.PHONY: secret
secret:
	openssl ecparam -genkey -name secp384r1 -out server.key
	openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650 -subj /C=US/ST=name/O=comp

.PHONY: install
install: mohawk
	install -D -m0755 mohawk $(DESTDIR)$(BINDIR)/mohawk

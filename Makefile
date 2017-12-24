PREFIX := $(GOPATH)
BINDIR := $(PREFIX)/bin
SOURCE := src/*.go src/*/*.go src/*/*/*.go

all: fmt mohawk

mohawk: $(SOURCE)
	go build -o mohawk src/*.go

.PHONY: fmt
fmt: $(SOURCE)
	gofmt -s -l -w $(SOURCE)

.PHONY: clean
clean:
	$(RM) mohawk

.PHONY: test
test:
	@echo "running smoke tests"
	bats test/mohawk.bats

.PHONY: test-unit
test-unit:
	@echo "running unit tests"
	@go test $(shell go list ./... | grep -v vendor)

.PHONY: secret
secret:
	openssl ecparam -genkey -name secp384r1 -out server.key
	openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650 -subj /C=US/ST=name/O=comp

.PHONY: container
container:
	# systemctl start docker
	docker build -t yaacov/mohawk ./
	docker tag yaacov/mohawk docker.io/yaacov/mohawk
	# docker push docker.io/yaacov/mohawk
	# docker run --name mohawk -e HAWKULAE_BACKEND="memory" -v $(readlink -f ./):/root/ssh:Z yaacov/mohawk

.PHONY: install
install: fmt mohawk
	install -D -m0755 mohawk $(DESTDIR)$(BINDIR)/mohawk

.PHONY: vendor
vendor:
	[ -d ${GOPATH}/src/github.com/LK4D4/vndr ] || go get -u -v github.com/LK4D4/vndr
	[ -d ${GOPATH}/src/github.com/MohawkTSDB/mohawk/vendor ] || ${GOPATH}/bin/vndr

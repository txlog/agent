.PHONY: all
all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean      Remove all artifacts"
	@echo "  fmt        Recursively format all packages"
	@echo "  vet        Recursively check all packages"
	@echo "  build      Compile a binary"
	@echo "  man        Compile the 'txlog' manpage"
	@echo "  rpm        Create the RPM package"

.PHONY: clean
clean:
	@rm -rf bin/
	@rm -rf doc/*.gz

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -ldflags="-s -w -extldflags=-static" -trimpath -o bin/txlog

.PHONY: man
man:
	@rm -f doc/txlog.1.gz
	@pandoc doc/txlog.1.md -s -t man -o doc/txlog.1
	@gzip doc/txlog.1
	@man -l doc/txlog.1.gz

.PHONY: rpm
rpm:
	@nfpm pkg --packager rpm --target ./bin/
	@rm -f ./bin/txlog

VERSION := $(shell cat .version)

.PHONY: all
all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean          Remove all artifacts"
	@echo "  fmt            Recursively format all packages"
	@echo "  vet            Recursively check all packages"
	@echo "  build          Compile a binary"
	@echo "  man            Compile the 'txlog' manpage"
	@echo "  rpm            Create the RPM package"
	@echo "  mcp-inspector  Run MCP inspector"

.PHONY: clean
clean:
	@rm -rf bin/
	@rm -rf man/*.gz

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: vet
vet:
	@go vet ./...

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" GOAMD64="v2" go build -ldflags="-s -w -extldflags=-static -X 'github.com/txlog/agent/cmd.agentVersion=$(VERSION)'" -trimpath -o bin/txlog

.PHONY: man
man:
	@rm -f man/txlog.1.gz
	@pandoc man/txlog.1.md -s -t man -o man/txlog.1 --metadata footer="txlog $(VERSION)" --metadata date="$(shell date '+%B %d, %Y')"
	@gzip man/txlog.1
	@man -l man/txlog.1.gz

.PHONY: rpm
rpm:
	@VERSION=$(VERSION) nfpm pkg --packager rpm --target ./bin/
	@rm -f ./bin/txlog

.PHONY: mcp-inspector
mcp-inspector:
	@npx @modelcontextprotocol/inspector

.PHONY: all clean build rpm man

all:
	@echo "Usage: make [OPTION]"
	@echo ""
	@echo "Options:"
	@echo "  clean      Remove all artifacts"
	@echo "  build      Compile a binary"
	@echo "  man        Compile the 'txlog' manpage"
	@echo "  rpm        Create the RPM package"

clean:
	@rm -rf bin/
	@rm -rf doc/*.gz

build:
	@GOOS="linux" GOARCH="amd64" go build -o bin/txlog

man:
	@rm -f doc/txlog.1.gz
	@pandoc doc/txlog.1.md -s -t man -o doc/txlog.1
	@gzip doc/txlog.1
	@man -l doc/txlog.1.gz

rpm:
	@nfpm pkg --packager rpm --target ./bin/

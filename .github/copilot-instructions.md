# Copilot Instructions for txlog/agent

## Repository Overview

**txlog/agent** is a Go-based command-line tool that compiles RPM package transaction data from yum/dnf and sends it to a central Txlog server for real-time monitoring and analytics. It helps system administrators track package history on RPM-based systems.

### Key Details
- **Language**: Go 1.24.3
- **CLI Framework**: Cobra (commands) + Viper (configuration)
- **Size**: Small focused codebase (~10 source files)
- **Target**: Linux RPM-based systems (RHEL, CentOS, Fedora)
- **Dependencies**: yum-utils, expects /etc/os-release and /etc/machine-id

## Build Instructions

### Prerequisites
**Always ensure these steps before building:**
```bash
go version  # Requires Go 1.24.3+
```

### Core Development Commands
**Run these commands in exact order for reliable builds:**

```bash
# 1. ALWAYS start with clean and format
make clean && make fmt && make vet

# 2. Build the binary
make build

# 3. Test the binary (requires config)
./bin/txlog --config conf/txlog.yaml version
```

### Command Reference
- `make clean` - Remove all artifacts (bin/, doc/*.gz)
- `make fmt` - Format Go code recursively (`go fmt ./...`)
- `make vet` - Static analysis (`go vet ./...`)  
- `make build` - Compile production binary to `./bin/txlog`
- `make man` - Generate man page (requires pandoc - see workarounds)
- `make rpm` - Create RPM package (requires nfpm - see workarounds)

### Build Workarounds

**Important**: Some build tools are optional for basic development:

1. **Missing pandoc** (for `make man`):
   - Error: `make: pandoc: No such file or directory`
   - Workaround: Skip man generation for code changes
   - Install if needed: Use the pandoc installation commands from README.md

2. **Missing nfpm** (for `make rpm`):
   - Error: `make: nfpm: No such file or directory`  
   - Workaround: Use `go build -o txlog .` directly for development
   - Install if needed: Use the nfpm installation commands from README.md

3. **Configuration Required**:
   - Binary fails without config: `Error reading config file: Config File "txlog" Not Found`
   - **Always use**: `./bin/txlog --config conf/txlog.yaml [command]`
   - Default config location: `/etc/txlog.yaml`

### Testing
**No automated tests exist** - validate manually:
```bash
# Test basic functionality
./bin/txlog --config conf/txlog.yaml --help
./bin/txlog --config conf/txlog.yaml version

# Test build command (will fail without actual dnf/yum data but validates binary)
./bin/txlog --config conf/txlog.yaml build

# Alternative build method (workaround)
go build -o test-binary . && ./test-binary --config conf/txlog.yaml version
```

**Expected outputs:**
- Version command: Shows "Txlog Agent v1.6.3" and "Txlog Server vunknown"
- Help: Shows available commands (build, help, version)
- Build command: Shows "Compiling host identification" then connection error (normal without server)

### Build Times
- `make build`: ~0.1-1 seconds (after initial go mod downloads)
- `go fmt ./...`: < 1 second
- `go vet ./...`: 1-3 seconds
- Initial dependency download: 10-30 seconds first time

## Project Architecture

### Directory Structure
```
/
├── main.go                    # Entry point - calls cmd.Execute()
├── cmd/                       # Cobra CLI commands
│   ├── root.go               # Root command + config initialization
│   ├── build.go              # Main functionality - transaction processing
│   └── version.go            # Version command + server communication
├── util/                     # Utility functions
│   ├── main.go               # Date conversion, machine ID, hostname, package parsing
│   ├── osrelease.go          # Parse /etc/os-release
│   └── needsRestarting.go    # Check if reboot needed (yum-utils)
├── conf/                     
│   └── txlog.yaml            # Default configuration file
├── doc/
│   └── txlog.1.md           # Man page source (markdown -> pandoc -> gzip)
├── Makefile                  # Build system
├── go.mod                    # Go modules
├── .goreleaser.yaml         # Release automation
└── nfpm.yaml                # RPM packaging config
```

### Configuration Files
- **Main config**: `conf/txlog.yaml` or `/etc/txlog.yaml`
- **Build config**: `Makefile`, `.goreleaser.yaml`, `nfpm.yaml`
- **Editor config**: `.editorconfig` (spaces, no tabs except Makefiles)
- **Dependencies**: `.github/dependabot.yml` (daily Go module updates)

### Key Dependencies
```go
// Core functionality
github.com/spf13/cobra       // CLI framework
github.com/spf13/viper       // Configuration management
github.com/go-resty/resty/v2 // HTTP client

// Utilities  
github.com/fatih/color       // Terminal colors
github.com/itlightning/dateparse // Date parsing
```

### Application Flow
1. **Configuration**: Load from `/etc/txlog.yaml` or `--config` flag
2. **Commands**:
   - `build`: Read dnf/yum transaction history → compile → send to server
   - `version`: Show agent/server versions + check for updates
3. **Server Communication**: REST API with optional basic auth

### Validation Pipeline
**No automated CI/CD exists** - manual validation only:
- Code formatting: `make fmt`
- Static analysis: `make vet`  
- Manual testing: Run binary with sample configs

### System Dependencies
- **Required**: `/etc/machine-id`, `/etc/os-release`
- **For functionality**: `yum-utils` package (needs-restarting command)
- **Build tools**: Go 1.24.3+
- **Optional**: pandoc (man pages), nfpm (RPM packaging)

## Important Implementation Notes

### Configuration Behavior
- **Critical**: Application exits with error if config file not found
- **Always specify config**: `--config conf/txlog.yaml` for development
- **Required config key**: `server.url` must be set
- **Optional**: `server.username`/`server.password` for auth

### Code Style
- **Indentation**: 2 spaces (enforced by .editorconfig)
- **Formatting**: Always run `make fmt` before committing
- **Error handling**: Functions return explicit errors, main exits on config errors

### File Dependencies
- **OS Release**: Reads `/etc/os-release` for system info
- **Machine ID**: Reads `/etc/machine-id` for unique identification  
- **Reboot Check**: Executes `needs-restarting -r` command

### API Integration
- **Server version**: GET `{server.url}/v1/version`
- **Executions**: POST `{server.url}/v1/executions`
- **Version check**: GET `https://txlog.rda.run/agent/version`

## Agent Instructions

**Trust these instructions** and only perform additional searches if information is incomplete or incorrect. Use the documented commands and workarounds to avoid common build failures.

**Key Success Patterns:**
1. Always use `--config conf/txlog.yaml` when testing
2. Run `make clean && make fmt && make vet` before building
3. Skip pandoc/nfpm steps if tools missing - focus on Go code
4. Validate changes by running `./bin/txlog --config conf/txlog.yaml version`
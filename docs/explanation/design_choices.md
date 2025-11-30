# Design Choices and Patterns

This document explains the technical choices and design patterns implemented in
the codebase.

## CLI Framework: Cobra & Viper

We use **Cobra** for command structure and **Viper** for configuration
management.

* **Why?**: This is the standard in the Go ecosystem (used by Kubernetes,
    Hugo, GitHub CLI). It provides a robust way to handle flags, config files,
    and environment variables seamlessly.
* **Pattern**: Commands are defined in the `cmd/` package, keeping `main.go`
    minimal.

## HTTP Client: Resty

We use **Resty** for HTTP communication.

* **Why?**: It simplifies common tasks like JSON marshaling/unmarshaling,
    setting headers, and handling retries. It reduces boilerplate code compared
    to the standard `net/http` library.

## Parsing Strategy: Regex vs. Libraries

The agent parses `dnf` output using Regular Expressions.

* **Why not use `libdnf`?**: Using C bindings (cgo) for `libdnf` would
    complicate the build process and cross-compilation. Parsing the CLI output
    allows the agent to be a static Go binary that works across different
    RPM-based distributions without dependency hell.
* **Trade-off**: This makes the agent sensitive to changes in DNF's output
    format, but provides greater portability.

## Error Handling

The application follows Go's explicit error handling pattern but adds context.

* **Pattern**: Errors are propagated up the stack and logged at the top level
    (usually in the `Run` function of the Cobra command).
* **User Feedback**: We distinguish between "system errors" (stack traces,
    debug logs) and "user errors" (configuration issues), presenting the latter
    with clear, colored messages.

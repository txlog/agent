---
trigger: always_on
---

# Conventions

## Version Management

- **Single source of truth**: `.version` file in root directory
- When bumping version: update `.version` and commit
- Server version compatibility checks are critical (e.g., API key auth requires server ≥1.14.0)

## Authentication Pattern

API key takes precedence over basic auth. All server requests use `util.SetAuthentication(request)`:

```go
// Prefer API key if set, fallback to basic auth
if apiKey := viper.GetString("server.api_key"); apiKey != "" {
    request.SetHeader("X-API-Key", apiKey)
} else if username := viper.GetString("server.username"); username != "" {
    request.SetBasicAuth(username, viper.GetString("server.password"))
}
```

## Package Manager Abstraction

Use `util.PackageBinary()` not hardcoded `yum`/`dnf`. It auto-detects based on OS version (RHEL/CentOS ≥8 uses DNF).

### Transaction Parsing Pattern

DNF output parsing uses regex with strict input validation:

```go
validInput := regexp.MustCompile(`^[a-zA-Z0-9_\-\./\\]+$`)
if !validInput.MatchString(transaction_id) {
    return TransactionDetail{}, fmt.Errorf("invalid input")
}
```

Always sanitize before exec.Command() to prevent injection.

## Date Handling

Use `util.DateConversion()` to convert DNF's varied date formats to RFC3339. Relies on `dateparse` library for fuzzy parsing.

## Error Handling in HTTP Clients

Check both network errors AND status codes:

```go
response, err := request.Post(url)
if err != nil {
    return 0, 0, err  // Network error
}
if response.StatusCode() != 200 {
    return 0, 0, fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
}
```

# TOML Configuration Loader for [Kong](https://github.com/alecthomas/kong)

Usage:

```go
kctx := kong.Parse(&cli, kong.Configuration(kongtoml.Loader, ".app.toml", "~/.app.toml"))
```

## Configuration mapping

Mapping is achieved by normalising both flags and configuration entries to
hyphen-separated keys.

eg. the flag `ftl init go --no-hermit` will be mapped to the configuration key
`init-go-no-hermit = true`, with fallback to `no-hermit = true`.

TOML sections are prefixed to attributes via concatenation with `-`. eg.

```toml
["init go"]
no-hermit = true
```

Will be mapped to `init-go-no-hermit = true`

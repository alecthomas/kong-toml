# TOML Configuration Loader for [Kong](https://github.com/alecthomas/kong)

Usage:

```go
kctx := kong.Parse(&cli, kong.Configuration(kongtoml.Loader, ".app.toml", "~/.app.toml"))
```

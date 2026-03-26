# mintlify-cli-sdk-go

Typed Go SDK for the [Mintlify CLI](https://mintlify.com/docs/development).

Wraps the Mintlify CLI as a subprocess with structured result types, pluggable npm runners, and first-class dev server lifecycle management.

## Install

```bash
go get github.com/albertocavalcante/mintlify-cli-sdk-go
```

Requires one of these npm executors on PATH: `mint`, `bunx`, `pnpm`, or `npx`.

## Usage

### Validate docs

```go
client, err := mintlify.New("./docs")
if err != nil {
    log.Fatal(err)
}

result, err := client.Validate(ctx, mintlify.ValidateOptions{Strict: true})
if err != nil {
    log.Fatal(err)
}

if result.OK {
    fmt.Println("All checks passed!")
} else {
    for _, e := range result.Errors {
        fmt.Printf("%s:%d:%d — %s\n", e.File, e.Line, e.Column, e.Message)
    }
}
```

### Build

```go
result, err := client.Build(ctx, mintlify.BuildOptions{})
if !result.OK {
    for _, e := range result.Errors {
        fmt.Printf("%s:%d — %s\n", e.File, e.Line, e.Message)
    }
}
fmt.Printf("Build took %s\n", result.Duration)
```

### Dev server

```go
server, err := client.StartDev(ctx, mintlify.DevOptions{Port: 3333})
if err != nil {
    log.Fatal(err)
}
defer server.Stop()

// Block until the server is accepting HTTP connections.
if err := server.WaitReady(ctx); err != nil {
    log.Fatal(err)
}

fmt.Printf("Preview ready at %s\n", server.URL())

// Wait for the server to exit (e.g. on Ctrl+C).
server.Wait()
```

### Check broken links

```go
result, err := client.BrokenLinks(ctx)
for _, link := range result.Links {
    fmt.Printf("%s (%d) in %s\n", link.URL, link.Status, link.Source)
}
```

## Available commands

| Method | CLI command | Result type |
|--------|------------|-------------|
| `Version()` | `mintlify version` | `VersionResult` |
| `Validate()` | `mintlify validate [--strict]` | `ValidateResult` |
| `Build()` | `mintlify build` | `BuildResult` |
| `BrokenLinks()` | `mintlify broken-links` | `BrokenLinksResult` |
| `A11y()` | `mintlify a11y` | `A11yResult` |
| `OpenAPICheck()` | `mintlify openapi-check [target]` | `OpenAPICheckResult` |
| `StartDev()` | `mintlify dev [--port N]` | `*DevServer` |
| `MigrateMDX()` | `mintlify migrate-mdx` | `string` |
| `Scrape()` | `mintlify scrape <mode> [target]` | `string` |
| `NewProject()` | `mintlify new [dir]` | `string` |
| `Rename()` | `mintlify rename` | `string` |
| `Upgrade()` | `mintlify upgrade` | `string` |

## Runner detection

The SDK auto-detects the best available npm executor in this priority order:

1. `mint` (system-wide Mintlify CLI)
2. `bunx mintlify`
3. `pnpm dlx mintlify`
4. `npx mintlify`

Override with `mintlify.WithRunner()`:

```go
client, _ := mintlify.New("./docs", mintlify.WithRunner(&mintlify.Runner{
    Name: "npx",
    Cmd:  "npx",
    Args: []string{"mintlify"},
}))
```

## Testing

Inject a mock command function to test without a real CLI:

```go
client, _ := mintlify.New("./docs",
    mintlify.WithRunner(&mintlify.Runner{Name: "mock", Cmd: "mock"}),
    mintlify.WithCommandFunc(func(ctx context.Context, dir, name string, args ...string) (string, string, int, error) {
        return "4.2.33\n", "", 0, nil
    }),
)
```

## License

MIT

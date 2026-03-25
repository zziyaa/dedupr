# dedupr

A macOS app for finding and removing duplicate files.

Notarized builds for macOS (Apple Silicon / arm64) are available on the [Releases](../../releases) page.

## Requirements

- Go 1.26+
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- Node.js + pnpm
- `jq` (`brew install jq`)

## Development

```sh
make dev      # Start dev server with live reload
make build    # Build release binary
make check    # Format, lint, and test
```

## License

MIT

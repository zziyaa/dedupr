# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**dedupr** is a macOS desktop application built with [Wails v2](https://wails.io/) — a Go backend exposing methods to a Svelte 5/TypeScript frontend rendered in a native WebView.

## Development Commands

### Quick Actions

```sh
make dev        # Start dev server with live reload (runs Wails dev mode)
make build      # Build release binary
make check      # Run fmt + lint + test (run before committing)
cd frontend && pnpm check   # TypeScript/Svelte type checking (frontend only)
```

## Architecture

### Go ↔ Frontend Bridge

- **`app.go`** is the Wails binding layer — all public methods on `App` become callable from the frontend
- Wails auto-generates TypeScript bindings into `frontend/wailsjs/go/` on each dev/build run — do not edit these files manually
- Backend → frontend events are emitted via `runtime.EventsEmit()`

### Frontend (`frontend/src/`)

- `App.svelte` — root component with custom draggable title bar
- `lib/components/MainPage.svelte` — main content area
- `lib/components/ui/` — reusable UI shadcn components (button, sonner toast wrapper)
- Backend calls: `import { MethodName } from '@wailsjs/go/main/App'`
- Backend events: `import { EventsOn } from '@wailsjs/runtime/runtime'`

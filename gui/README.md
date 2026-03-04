# tp-gui

Cross-platform GUI for [tp](https://github.com/dfairburn/tp) - the template HTTP client.

## Features

- **3-panel layout** matching the TUI experience:
  - Template browser (left) - Navigate your templates with folder tree structure
  - Request panel (top-right) - View URL, method, headers, and body
  - Response panel (bottom-right) - View response with syntax highlighting

- **Cross-platform**: Runs on macOS, Windows, and Linux
- **Native performance**: Built with Go backend, small binary (~15MB)
- **Keyboard shortcuts**: 
  - `Cmd/Ctrl + Enter` - Execute current request
  - `Escape` - Clear selection
- **Dark/Light theme**: Auto-detects system preference

## Prerequisites

- Go 1.21+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Development

Run in live development mode with hot reload:

```bash
cd gui
wails dev
```

This starts a Vite dev server with hot reload for frontend changes.

## Building

### Current Platform

```bash
cd gui
wails build
```

The built application will be in `build/bin/`.

### Cross-platform Builds

Build for all platforms (requires appropriate cross-compilation setup):

```bash
# macOS (Intel)
wails build -platform darwin/amd64

# macOS (Apple Silicon)
wails build -platform darwin/arm64

# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64
```

## Project Structure

```
gui/
├── app.go              # Go backend - exposes tp handlers to frontend
├── main.go             # Wails application entry point
├── frontend/
│   ├── src/
│   │   ├── App.svelte            # Main application component
│   │   ├── lib/
│   │   │   ├── components/       # Svelte components
│   │   │   │   ├── TemplateList.svelte
│   │   │   │   ├── RequestPanel.svelte
│   │   │   │   └── ResponsePanel.svelte
│   │   │   └── stores/           # Svelte stores for state
│   │   │       └── app.ts
│   │   └── style.css
│   └── wailsjs/                  # Auto-generated TypeScript bindings
└── build/
    └── bin/                      # Built applications
```

## License

MIT

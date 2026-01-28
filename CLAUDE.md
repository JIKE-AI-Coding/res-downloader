# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**res-downloader** is a cross-platform resource downloader built with Go (Wails v2) + Vue.js frontend. It acts as a local proxy server that intercepts network traffic to detect and download media resources (videos, audio, images) from platforms like WeChat, TikTok, etc.

## Build & Development Commands

### Prerequisites
- Go 1.22+
- Node.js (for frontend)
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Development
```bash
# Install frontend dependencies
cd frontend && npm install

# Development mode (hot reload)
wails dev

# Build for current platform
wails build

# Build for specific platforms
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform darwin/arm64
wails build -platform linux/amd64
```

### Frontend Only
```bash
cd frontend
npm run dev     # Start Vite dev server
npm run build   # Build frontend (TypeScript check + Vite build)
```

### Running Tests
```bash
# Run Go tests
go test ./...

# Run Go tests with coverage
go test -cover ./...
```

## Architecture

### High-Level Design

The application uses a **proxy-based architecture** for resource interception:

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Browser   │────▶│  HTTP Server │────▶│   Plugins   │
│  (External) │     │  (127.0.0.1) │     │ (Handlers)  │
└─────────────┘     └──────────────┘     └─────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   Resource   │
                     │   Detector   │
                     └──────────────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   Frontend   │
                     │  (Vue.js)    │
                     └──────────────┘
```

### Core Components

**main.go** - Entry point that initializes Wails app with embedded frontend assets

**core/app.go** - Singleton App instance managing:
- Certificate installation (for HTTPS MITM)
- System proxy configuration (platform-specific)
- App lifecycle (startup/exit)
- Global singletons for all subsystems

**core/proxy.go** - HTTP/HTTPS proxy server using `goproxy`:
- MITM (Man-in-the-Middle) TLS interception
- Plugin registry for domain-specific handlers
- Route requests to plugins based on top-level domain
- Conditional MITM based on RuleSet (which domains to decrypt)

**core/http.go** - Internal HTTP API server:
- Shares the same port as proxy (127.0.0.1:8899)
- Handles API endpoints for frontend communication
- Proxy traffic goes to `proxyOnce.Proxy.ServeHTTP()`
- API calls handled by `HandleApi()`

**core/downloader.go** - Multi-threaded file downloader:
- Concurrent chunk-based downloads with Range requests
- Retry logic with fallback to single-threaded mode
- Progress callbacks via channels
- Context-based cancellation support

**core/config.go** - Configuration management:
- JSON-based storage with default fallback
- MIME type mapping (ContentType → classification + extension)
- Domain rules for MITM filtering
- Merges user config with defaults on load

**core/rule.go** - Domain matching for MITM decisions:
- Wildcard support (`*.example.com`)
- Negation rules (`!example.com`)
- Global wildcard (`*`)
- Determines which HTTPS connections get decrypted

**core/resource.go** - Resource detection and management:
- Deduplicates resources by URL MD5 signature
- Tracks marked resources to avoid duplicates
- Sends new resources to frontend via WebSocket events

**core/system*.go** - Platform-specific system proxy:
- `system_windows.go` - Windows registry proxy settings
- `system_darwin.go` - macOS network settings
- `system_linux.go` - Linux environment/GSettings

**core/plugins/** - Extensible plugin system:
- **Plugin Interface** (`core/shared/plugin.go`):
  - `Domains() []string` - Register domains to handle
  - `OnRequest()` - Modify outgoing requests
  - `OnResponse()` - Intercept responses for resource detection
- **Bridge Pattern** - Plugins access core via callback interface (avoid circular imports)
- **DefaultPlugin** - Generic handler matching by Content-Type
- **QqPlugin** - Example domain-specific handler

**core/storage.go** - Simple file-based persistence layer

**core/middleware.go** - Wails asset server middleware for embedding

### Frontend Structure

**frontend/src/api/** - Wails bindings:
- Auto-generated from Go structs in `wailsjs/go/core/`
- `request.ts` - Axios-based HTTP client to 127.0.0.1:8899

**frontend/src/views/** - Main pages:
- `index.vue` - Resource list with download controls
- `setting.vue` - Configuration UI

**frontend/src/stores/** - Pinia state management
**frontend/src/components/** - Reusable Vue components
**frontend/src/locales/** - i18n translations (zh/en)

### Key Data Flow

1. **Resource Detection:**
   - Browser → System Proxy → This App (127.0.0.1:8899)
   - `proxy.go` routes to plugin based on domain
   - Plugin's `OnResponse()` examines Content-Type
   - If matches configured type → Create `MediaInfo` → Send to frontend via `runtime.EventsEmit()`

2. **Download:**
   - Frontend calls `/download` API with `MediaInfo`
   - `resource.go` creates `FileDownloader` with progress callback
   - Downloader spawns goroutines per chunk
   - Progress sent via `httpServerOnce.send()` (WebSocket events)

3. **Configuration:**
   - Stored in user config directory: `~/.config/res-downloader/`
   - `config.json` persisted on change
   - `install.lock` tracks certificate installation

## Singleton Pattern

The codebase uses global singletons (initialized in `app.go:GetApp()`):
- `appOnce` - Main app instance
- `globalConfig` - Configuration
- `globalLogger` - Logging
- `resourceOnce` - Resource manager
- `systemOnce` - System proxy operations
- `proxyOnce` - HTTP/HTTPS proxy
- `httpServerOnce` - Internal API server
- `ruleOnce` - Domain rules engine

## Certificate & HTTPS MITM

The app embeds a self-signed CA certificate (`app.go:58-110`) for:
- Decrypting HTTPS traffic (MITM)
- Installing to system trust store on first run
- Platform-specific installation in `system_*.go` files

## Adding New Plugins

To add a new domain-specific handler:

1. Create file in `core/plugins/` (e.g., `plugin.example.com.go`)
2. Implement `shared.Plugin` interface:
   ```go
   type ExamplePlugin struct {
       bridge *shared.Bridge
   }
   func (p *ExamplePlugin) SetBridge(bridge *shared.Bridge) { p.bridge = bridge }
   func (p *ExamplePlugin) Domains() []string { return []string{"example.com"} }
   func (p *ExamplePlugin) OnRequest(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
       return r, nil // Pass through
   }
   func (p *ExamplePlugin) OnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
       // Detect and extract resources
       // Call p.bridge.Send("newResources", mediaInfo) when found
       return resp
   }
   ```
3. Register in `core/proxy.go:init()` plugin list

## Important Notes

- **Version parsing**: App version extracted from `wails.json` via regex in `app.go:45`
- **Port**: Default proxy port is 8899 (configurable)
- **Thread safety**: Config uses `sync.RWMutex` for MIME map access
- **Forbidden headers**: `downloader.go:85-107` lists headers stripped from download requests
- **Retry logic**: Downloads retry up to 3 times with 3s delay, fallback to single-threaded on failure

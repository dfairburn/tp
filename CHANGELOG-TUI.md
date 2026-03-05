# TUI Feature Changes

This document summarizes the TUI features added during the development session.

## Features Added

### 1. Syntax Highlighting (Chroma)

**Files:** `tui/styles.go`, `tui/model.go`

Added syntax highlighting for request and response bodies using the Chroma library.

- **Response body**: Highlighted based on Content-Type (JSON, XML, HTML, YAML, GraphQL)
- **Request body**: Detects Go template tags (`{{...}}`) and uses appropriate lexer
- Uses Monokai theme with Terminal256 formatter

**Key functions in `tui/styles.go`:**
- `highlightCode(input, lexerName string)` - Basic syntax highlighting
- `getLexerForContentType(contentType string)` - Maps Content-Type to lexer name
- `highlightCodeWithTemplates(input, lexerName string)` - Handles Go template syntax

**Dependency:** `github.com/alecthomas/chroma/v2`

---

### 2. Response Caching Per Template

**File:** `tui/model.go`

Responses are cached in memory by template path, so switching between templates preserves their responses.

- `responseCache map[string]*RequestResult` field on Model
- Cache populated when request completes (`requestCompleteMsg` handler)
- Cache loaded when selecting a template (Enter key or search result)

---

### 3. Session Persistence

**File:** `tui/model.go`

Session state is persisted to `~/.tp/session.json` and restored on startup.

**What's persisted:**
- Selected template path
- All cached responses

**Structs:**
- `SessionCache` - Top-level JSON structure
- `CachedResponse` - JSON-serializable version of `RequestResult`

**Key functions:**
- `sessionCacheFilePath()` - Returns `~/.tp/session.json`
- `saveCache()` - Called on quit (q/ctrl+c)
- `loadCache()` - Called on startup

**Restoration flow:**
1. `loadCache()` reads session.json and stores `pendingSelectedTemplate`
2. After templates load (`templatesLoadedMsg`), the selected template is restored
3. Parent folders are expanded and cursor positioned on the template
4. Cached response is displayed

---

### 4. Improved Filter/Search UX

**File:** `tui/model.go`

Redesigned the search to work as a persistent filter.

**Workflow:**
1. Press `/` to start filtering
2. Type filter text (list filters in real-time)
3. Press `Enter`, `Up`, or `Down` to lock in the filter
4. Navigate filtered results with normal keys (j/k, arrows, ctrl+u/d)
5. Press `Enter` to select a template
6. Press `/` again to edit or clear the filter

**Behavior:**
- Filter bar stays visible when filter is active (shows `/yourfilter`)
- `Esc` cancels and clears the filter
- `Up`/`Down` while typing locks the filter and moves cursor

---

## File Summary

| File | Changes |
|------|---------|
| `tui/model.go` | Response caching, session persistence, filter UX, template restoration |
| `tui/styles.go` | Chroma syntax highlighting functions |
| `tui/templates.go` | Added `FindByPath()` method |
| `go.mod` / `go.sum` | Added chroma dependency |

---

## Key Code Locations

- **Session cache structs:** `tui/model.go` ~lines 599-617
- **Save/load cache:** `tui/model.go` ~lines 633-700
- **Template restoration:** `tui/model.go` ~lines 207-222 (in `templatesLoadedMsg` handler)
- **Filter mode handler:** `tui/model.go` ~lines 513-540
- **Syntax highlighting:** `tui/styles.go` ~lines 256-330

# tp Feature Comparison

This document tracks feature parity across all three versions of tp: Command Line (CLI), Terminal User Interface (TUI), and Graphical User Interface (GUI).

## Feature Matrix

| Feature | CLI | TUI | GUI | Notes |
|---------|-----|-----|-----|-------|
| **Template Management** |
| List templates | ✅ | ✅ | ✅ | CLI: tree view; TUI/GUI: interactive browser |
| Create template | ✅ | ✅ | ✅ | CLI/TUI: opens `$EDITOR`; GUI: visual form editor |
| Edit template | ✅ | ✅ | ✅ | CLI/TUI: opens `$EDITOR`; GUI: visual editor or external |
| Delete template | ❌ | ✅ | ✅ | |
| Rename template | ❌ | ❌ | ✅ | |
| Move/organize templates | ❌ | ❌ | ✅ | GUI: drag-and-drop support |
| Duplicate template | ❌ | ❌ | ❌ | Can be done via save-as |
| Create folders | ❌ | ❌ | ✅ | |
| Search/filter templates | ✅ | ✅ | ✅ | CLI: fuzzy finder; TUI: `/` key; GUI: search box |
| Nested template directories | ✅ | ✅ | ✅ | All versions support folder organization |
| **Request Execution** |
| Execute HTTP requests | ✅ | ✅ | ✅ | |
| Display request details | ✅ | ✅ | ✅ | Headers, body, method, URL |
| Display response | ✅ | ✅ | ✅ | |
| Response headers view | ✅ | ✅ | ✅ | |
| Response body view | ✅ | ✅ | ✅ | |
| Response status code | ✅ | ✅ | ✅ | TUI/GUI: color-coded |
| Response time | ❌ | ✅ | ✅ | |
| Response size | ❌ | ✅ | ✅ | |
| Syntax highlighting | ❌ | ✅ | ✅ | JSON, XML, GraphQL support |
| JSON formatting | ✅ | ✅ | ✅ | CLI: via `jq` or similar; TUI/GUI: built-in |
| Raw output mode | ✅ | ❌ | ❌ | CLI: `--raw` flag |
| **Variable Management** |
| Environment variables | ✅ | ✅ | ✅ | YAML-based variable files |
| Nested variables | ✅ | ✅ | ✅ | e.g., `.users.user_1` |
| Shell command expansion | ✅ | ✅ | ✅ | e.g., `token: $(get-token)` |
| Variable overrides | ✅ | ✅ | ✅ | CLI: `-o` flag; TUI/GUI: persistent file |
| Persistent overrides | ❌ | ✅ | ✅ | Stored in `~/.tp/overrides.yaml` |
| Edit overrides file | ✅ | ✅ | ✅ | CLI: `tp env`; TUI: `o` key; GUI: settings panel |
| Inline variable editing | ❌ | ❌ | ✅ | GUI: edit overrides in UI |
| View merged variables | ❌ | ❌ | ✅ | GUI: shows env + overrides with sources |
| Refresh variables | ❌ | ❌ | ✅ | GUI: re-executes shell commands |
| Variable auto-completion | ✅ | ❌ | ❌ | CLI: shell completion |
| **Template Functions** |
| `default` function | ✅ | ✅ | ✅ | Provide default values |
| `optional` function | ✅ | ✅ | ✅ | Conditional content inclusion |
| `timestamp` function | ✅ | ✅ | ✅ | RFC3339 timestamps with duration support |
| **User Interface** |
| Color-coded methods | ❌ | ✅ | ✅ | Visual HTTP method indicators |
| Keyboard shortcuts | ✅ | ✅ | ✅ | CLI: shell completion; TUI: vim-like; GUI: Cmd/Ctrl |
| Mouse support | ❌ | ✅ | ✅ | |
| Tabbed interface | ❌ | ✅ | ✅ | Request/response sections |
| Help overlay | ❌ | ✅ | ❌ | TUI: `?` key |
| Status bar | ❌ | ✅ | ❌ | TUI: shows mode and hints |
| Copy to clipboard | ❌ | ✅ | ✅ | TUI: `y` key; GUI: copy button |
| Theme support | ❌ | ❌ | ✅ | GUI: light/dark with auto-detection |
| Loading indicators | ❌ | ✅ | ✅ | Visual feedback during requests |
| **Session & State** |
| Response caching | ❌ | ✅ | ✅ | TUI: persists across sessions; GUI: per-template |
| Last template restoration | ❌ | ✅ | ❌ | TUI: restores last selected template |
| Session persistence | ❌ | ✅ | ❌ | TUI: `~/.tp/session.json` |
| Settings persistence | ❌ | ❌ | ✅ | GUI: theme, window size |
| **Configuration** |
| Initialize config | ✅ | ✅ | ✅ | CLI: `tp init`; TUI/GUI: auto-created |
| Edit config file | ✅ | ❌ | ✅ | CLI: `tp config`; GUI: settings panel |
| Custom templates directory | ✅ | ✅ | ✅ | All: `~/.tp/config.yml` |
| Custom environment file | ✅ | ✅ | ✅ | All: configurable path |
| Multiple environment files | ✅ | ❌ | ❌ | CLI: `-e` flag |
| Config overrides (flags) | ✅ | ❌ | ❌ | CLI: `--config`, `--vars` flags |
| Debug logging | ✅ | ❌ | ❌ | CLI: `--debug` flag |
| Custom log file | ✅ | ❌ | ❌ | CLI: `--log` flag |
| **Migration & Integration** |
| Insomnia migration | ✅ | ❌ | ❌ | CLI: `tp migrate insomnia` |
| Postman migration | ✅ | ❌ | ❌ | CLI: `tp migrate postman` |
| Shell completion | ✅ | ❌ | ❌ | CLI: bash, zsh, fish, powershell |
| Template help docs | ✅ | ❌ | ❌ | CLI: shows template descriptions |
| **Advanced Features** |
| External editor integration | ✅ | ✅ | ✅ | All: respects `$EDITOR` environment variable |
| File safety checks | ❌ | ❌ | ✅ | GUI: prevents unsafe operations |
| Drag-and-drop | ❌ | ❌ | ✅ | GUI: reorganize templates |
| Error display | ✅ | ✅ | ✅ | All: show detailed error messages |
| Cross-platform support | ✅ | ✅ | ✅ | All: macOS, Linux, Windows |

## Legend

- ✅ = Fully supported
- ⚠️ = Partially supported or limited
- ❌ = Not supported
- 🚧 = Planned/in development

## Version-Specific Commands

### CLI Only
- `tp init` - Initialize configuration
- `tp completion` - Generate shell completions
- `tp migrate` - Import from other API tools
- `tp list` - List templates in tree format
- `tp config` - Open config file
- `tp env` - Open environment file

### TUI Only
Launched via `tp tui` command. Interactive keybindings:
- `j/k` or arrows - Navigate
- `enter` - Select template
- `space` - Toggle folder
- `x` or `ctrl+enter` - Execute request
- `n` - New template
- `e` - Edit template
- `d` - Delete template
- `o` - Edit overrides
- `r` - Refresh templates
- `y` - Copy response to clipboard
- `/` - Search/filter
- `?` - Toggle help
- `tab` - Switch panels
- `q` - Quit

### GUI Only
Native application with:
- Visual template editor (forms instead of YAML)
- Settings panel for variables and overrides
- Drag-and-drop template organization
- Theme switcher (light/dark mode)
- Inline variable override editing
- One-click file operations (create folder, rename, move)

## Implementation Status

### Feature Parity Goals

#### High Priority (TUI → GUI sync)
- [ ] Session persistence in GUI (restore last template)
- [ ] Help overlay in GUI
- [ ] Status bar with contextual hints in GUI

#### Medium Priority
- [ ] Template duplication in all versions
- [ ] Delete templates in CLI
- [ ] Rename templates in CLI/TUI
- [ ] Migration tools in GUI

#### Low Priority
- [ ] Raw output mode in TUI/GUI
- [ ] Multiple environment files in TUI/GUI
- [ ] Debug logging in TUI/GUI

## Development Notes

### CLI Version
- **Location**: `cmd/tp/`
- **Entry point**: `cmd/tp/main.go`
- **Best for**: Scripting, automation, CI/CD pipelines
- **Unique strengths**: Shell completion, migration tools, flags for automation

### TUI Version
- **Location**: `tui/`
- **Entry point**: `tui/tui.go`
- **Launch**: `tp tui`
- **Best for**: Interactive terminal work, SSH sessions, keyboard-driven workflows
- **Unique strengths**: Session persistence, cached responses, vim-like navigation

### GUI Version
- **Location**: `gui/`
- **Technology**: Wails v2 (Go + Svelte)
- **Build**: `wails build` or `wails dev`
- **Best for**: Visual editing, beginners, drag-and-drop organization
- **Unique strengths**: Visual editor, theme support, inline editing

## Contributing

When adding new features:
1. Consider implementing in all three versions for feature parity
2. Update this document with the new feature status
3. Add tests for each version
4. Update relevant documentation (README, help text, etc.)

## Version History

- **v0.1.0** - Initial CLI release
- **v0.2.0** - Added TUI version
- **v0.3.0** - Added GUI version (current)

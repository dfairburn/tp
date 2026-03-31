package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/paths"
	"github.com/dfairburn/tp/static"
	logging "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Panel represents the focused panel
type Panel int

const (
	PanelTemplates Panel = iota
	PanelRequest
	PanelResponse
)

// Tab in request/response panel
type Tab int

const (
	TabHeaders Tab = iota
	TabBody
	TabParams
)

// Mode represents the current input mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeNewTemplate
	ModeConfirmDelete
)

// Message types
type templatesLoadedMsg struct {
	err error
}

type requestCompleteMsg struct {
	result       *RequestResult
	templatePath string
}

type editorFinishedMsg struct {
	err error
}

type clearStatusMsg struct{}

// Model is the main TUI model
type Model struct {
	// Config
	config  config.Config
	envFile string
	vars    map[interface{}]interface{}
	logger  *logging.Logger

	// Panels
	activePanel Panel
	mode        Mode

	// Template list
	templates *TemplateList

	// Request details
	selectedTemplate *TemplateItem
	requestTab       Tab

	// Response
	response      *RequestResult
	responseTab   Tab
	responseCache map[string]*RequestResult // Cache responses by template path

	// Viewports for scrollable content
	requestViewport  viewport.Model
	responseViewport viewport.Model

	// Inputs
	searchInput textinput.Model

	// Overrides
	overrides []config.Override

	// State
	loading                 bool
	spinner                 spinner.Model
	err                     error
	pendingSelectedTemplate string // Template path to restore after templates load

	// New template input
	newTemplateInput textinput.Model

	// Delete confirmation
	pendingDeleteTemplate *TemplateItem

	// Dimensions
	width  int
	height int

	// Help visibility
	showHelp bool

	// Status message (temporary feedback)
	statusMsg string
}

// New creates a new TUI model
func New(logger *logging.Logger, cfg config.Config, envFile string) Model {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search templates..."
	searchInput.CharLimit = 50

	// Initialize new template input
	newTemplateInput := textinput.New()
	newTemplateInput.Placeholder = "template-name"
	newTemplateInput.CharLimit = 100

	// Initialize viewports
	requestVP := viewport.New(0, 0)
	responseVP := viewport.New(0, 0)

	// Load environment
	_, vars := config.LoadEnvironment(logger, envFile, cfg.EnvironmentFile)

	m := Model{
		config:           cfg,
		envFile:          envFile,
		vars:             vars,
		logger:           logger,
		activePanel:      PanelTemplates,
		mode:             ModeNormal,
		templates:        NewTemplateList(cfg.TemplatesDirectoryPath),
		requestTab:       TabBody,
		responseTab:      TabBody,
		requestViewport:  requestVP,
		responseViewport: responseVP,
		searchInput:      searchInput,
		newTemplateInput: newTemplateInput,
		overrides:        make([]config.Override, 0),
		responseCache:    make(map[string]*RequestResult),
		spinner:          s,
		showHelp:         true,
	}

	// Load persisted overrides
	m.loadOverrides()

	// Load persisted response cache
	m.loadCache()

	return m
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadTemplates,
		m.spinner.Tick,
	)
}

func (m Model) loadTemplates() tea.Msg {
	err := m.templates.Load(m.logger)
	return templatesLoadedMsg{err: err}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewportSizes()
		return m, nil

	case templatesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		}
		// Restore selected template from cache
		if m.pendingSelectedTemplate != "" {
			item := m.templates.FindByPath(m.pendingSelectedTemplate)
			if item != nil {
				m.selectedTemplate = item
				m.templates.selectedItem = item
				// Position cursor on the item (expands parent folders if needed)
				m.templates.ClearSearchAndFocus(item)
				m.requestViewport.SetContent(m.renderRequestDetails())
				// Load cached response if available
				if cached, ok := m.responseCache[m.pendingSelectedTemplate]; ok {
					m.response = cached
					m.responseViewport.SetContent(m.renderResponseBody())
				}
			}
			m.pendingSelectedTemplate = ""
		}
		return m, nil

	case requestCompleteMsg:
		m.loading = false
		m.response = msg.result
		// Cache the response for this template
		if msg.templatePath != "" {
			m.responseCache[msg.templatePath] = msg.result
		}
		m.responseViewport.SetContent(m.renderResponseBody())
		m.responseViewport.GotoTop()
		return m, nil

	case editorFinishedMsg:
		// Reload templates and overrides after editor closes
		if msg.err != nil {
			m.err = msg.err
		}
		// Reload overrides in case the overrides file was edited
		m.loadOverrides()
		// Reload the selected template if it was being edited
		if m.selectedTemplate != nil {
			_ = m.selectedTemplate.LoadMetadata()
			m.requestViewport.SetContent(m.renderRequestDetails())
		}
		return m, m.loadTemplates

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil
	}

	// Update sub-components
	if m.mode == ModeSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		cmds = append(cmds, cmd)
		m.templates.Filter(m.searchInput.Value())
	}

	if m.mode == ModeNewTemplate {
		var cmd tea.Cmd
		m.newTemplateInput, cmd = m.newTemplateInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewports
	if m.activePanel == PanelRequest {
		var cmd tea.Cmd
		m.requestViewport, cmd = m.requestViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.activePanel == PanelResponse {
		var cmd tea.Cmd
		m.responseViewport, cmd = m.responseViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "ctrl+c", "q":
		if m.mode == ModeNormal {
			m.saveCache()
			return m, tea.Quit
		}
		// Exit current mode
		m.mode = ModeNormal
		m.searchInput.Blur()
		return m, nil

	case "esc":
		if m.mode != ModeNormal {
			m.mode = ModeNormal
			m.searchInput.Blur()
			return m, nil
		}
		return m, nil

	case "?":
		if m.mode == ModeNormal {
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	// Mode-specific handling
	switch m.mode {
	case ModeSearch:
		return m.handleSearchMode(msg)
	case ModeNewTemplate:
		return m.handleNewTemplateMode(msg)
	case ModeConfirmDelete:
		return m.handleConfirmDeleteMode(msg)
	default:
		return m.handleNormalMode(msg)
	}
}

func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	// Panel navigation (3 panels: Templates, Request, Response)
	case "tab":
		m.activePanel = (m.activePanel + 1) % 3
		return m, nil
	case "shift+tab":
		m.activePanel = (m.activePanel + 2) % 3 // Wrap backwards
		return m, nil
	case "1":
		m.activePanel = PanelTemplates
		return m, nil
	case "2":
		m.activePanel = PanelRequest
		return m, nil
	case "3":
		m.activePanel = PanelResponse
		return m, nil

	// Search mode
	case "/":
		m.mode = ModeSearch
		m.searchInput.Focus()
		return m, textinput.Blink

	// Override mode - open overrides file in editor
	case "o":
		return m, m.openOverridesInEditor()

	// Execute request
	case "enter":
		if m.activePanel == PanelTemplates {
			selected := m.templates.Select()
			if selected != nil {
				m.selectedTemplate = selected
				m.requestViewport.SetContent(m.renderRequestDetails())
				m.requestViewport.GotoTop()
				// Load cached response if available
				if cached, ok := m.responseCache[selected.AbsolutePath]; ok {
					m.response = cached
				} else {
					m.response = nil
				}
				m.responseViewport.SetContent(m.renderResponseBody())
				m.responseViewport.GotoTop()
			}
			return m, nil
		}
		return m, nil

	case "x", "ctrl+enter":
		if m.selectedTemplate != nil && !m.loading {
			m.loading = true
			m.response = nil
			return m, m.executeRequest
		}
		return m, nil

	// Refresh templates
	case "r":
		if m.activePanel == PanelTemplates {
			m.loading = true
			return m, m.loadTemplates
		}
		return m, nil

	// Edit current template
	case "e":
		if m.selectedTemplate != nil {
			return m, m.openEditor(m.selectedTemplate.AbsolutePath)
		}
		return m, nil

	// Create new template
	case "n":
		m.mode = ModeNewTemplate
		m.newTemplateInput.SetValue("")
		m.newTemplateInput.Focus()
		return m, textinput.Blink

	// Copy response to clipboard
	case "y":
		if m.response != nil && m.response.Body != "" {
			err := clipboard.WriteAll(m.response.Body)
			if err != nil {
				m.statusMsg = "Failed to copy: " + err.Error()
			} else {
				m.statusMsg = "Copied response to clipboard"
			}
		} else {
			m.statusMsg = "No response to copy"
		}
		return m, m.clearStatusAfterDelay()

	// Delete template
	case "d":
		if m.activePanel == PanelTemplates {
			current := m.templates.Current()
			if current != nil && !current.IsDir {
				m.pendingDeleteTemplate = current
				m.mode = ModeConfirmDelete
				return m, nil
			}
		}
		return m, nil
	}

	// Panel-specific keys
	switch m.activePanel {
	case PanelTemplates:
		return m.handleTemplatesPanelKeys(msg)
	case PanelRequest:
		return m.handleRequestPanelKeys(msg)
	case PanelResponse:
		return m.handleResponsePanelKeys(msg)
	}

	return m, nil
}

func (m Model) handleTemplatesPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Calculate page size based on visible template list height
	panelHeight := m.height - 4
	listHeight := panelHeight - 4
	if m.mode == ModeSearch {
		listHeight -= 3
	}
	pageSize := max(1, listHeight-1)

	switch msg.String() {
	case "j", "down":
		m.templates.MoveDown()
	case "k", "up":
		m.templates.MoveUp()
	case "ctrl+d", "pgdown":
		m.templates.PageDown(pageSize)
	case "ctrl+u", "pgup":
		m.templates.PageUp(pageSize)
	case "space":
		m.templates.ToggleExpand()
	case "g":
		m.templates.cursor = 0
	case "G":
		m.templates.cursor = m.templates.Len() - 1
	}
	return m, nil
}

func (m Model) handleRequestPanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.requestViewport.ScrollDown(1)
	case "k", "up":
		m.requestViewport.ScrollUp(1)
	case "h":
		m.requestTab = TabHeaders
		m.requestViewport.SetContent(m.renderRequestDetails())
	case "b":
		m.requestTab = TabBody
		m.requestViewport.SetContent(m.renderRequestDetails())
	case "p":
		m.requestTab = TabParams
		m.requestViewport.SetContent(m.renderRequestDetails())
	}
	return m, nil
}

func (m Model) handleResponsePanelKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		m.responseViewport.ScrollDown(1)
	case "k", "up":
		m.responseViewport.ScrollUp(1)
	case "d":
		m.responseViewport.HalfPageDown()
	case "u":
		m.responseViewport.HalfPageUp()
	case "g":
		m.responseViewport.GotoTop()
	case "G":
		m.responseViewport.GotoBottom()
	case "h":
		m.responseTab = TabHeaders
		m.responseViewport.SetContent(m.renderResponseBody())
	case "b":
		m.responseTab = TabBody
		m.responseViewport.SetContent(m.renderResponseBody())
	}
	return m, nil
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "down", "up":
		// Lock in the filter and return to normal mode
		m.mode = ModeNormal
		m.searchInput.Blur()
		// If up/down, also move in that direction
		if msg.String() == "down" {
			m.templates.MoveDown()
		} else if msg.String() == "up" {
			m.templates.MoveUp()
		}
		return m, nil
	case "esc":
		// Cancel filter and return to normal mode
		m.templates.Filter("")
		m.searchInput.SetValue("")
		m.mode = ModeNormal
		m.searchInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.templates.Filter(m.searchInput.Value())
	return m, cmd
}

// overridesFilePath returns the path to the overrides persistence file
// Stored in ~/.tp/ alongside the templates folder
func (m Model) overridesFilePath() string {
	return filepath.Join(config.DefaultDirectory, "overrides.yaml")
}

// saveOverrides persists the current overrides to disk
func (m Model) saveOverrides() {
	data := make(map[string]string)
	for _, o := range m.overrides {
		data[o.Key] = o.Value
	}

	content, err := yaml.Marshal(data)
	if err != nil {
		m.logger.Warnf("failed to marshal overrides: %v", err)
		return
	}

	err = os.WriteFile(m.overridesFilePath(), content, 0644)
	if err != nil {
		m.logger.Warnf("failed to save overrides: %v", err)
	}
}

// loadOverrides loads persisted overrides from disk
func (m *Model) loadOverrides() {
	content, err := os.ReadFile(m.overridesFilePath())
	if err != nil {
		// File doesn't exist or can't be read - that's fine
		return
	}

	data := make(map[string]string)
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		m.logger.Warnf("failed to parse overrides file: %v", err)
		return
	}

	m.overrides = make([]config.Override, 0, len(data))
	for k, v := range data {
		m.overrides = append(m.overrides, config.Override{Key: k, Value: v})
	}
}

// CachedResponse is a JSON-serializable version of RequestResult
type CachedResponse struct {
	StatusCode  int               `json:"status_code"`
	Status      string            `json:"status"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	DurationMs  int64             `json:"duration_ms"`
	ContentType string            `json:"content_type"`
	Size        int               `json:"size"`
	Request     *RequestInfo      `json:"request,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// SessionCache is the top-level structure for the cache file
type SessionCache struct {
	SelectedTemplate string                     `json:"selected_template,omitempty"`
	Responses        map[string]*CachedResponse `json:"responses,omitempty"`
}

// sessionCacheFilePath returns the path to the session cache file
func (m Model) sessionCacheFilePath() string {
	return filepath.Join(config.DefaultDirectory, "session.json")
}

// saveCache persists the response cache and selected template to disk
func (m Model) saveCache() {
	session := SessionCache{
		Responses: make(map[string]*CachedResponse),
	}

	if m.selectedTemplate != nil {
		session.SelectedTemplate = m.selectedTemplate.AbsolutePath
	}

	for path, result := range m.responseCache {
		if result == nil {
			continue
		}
		cached := &CachedResponse{
			StatusCode:  result.StatusCode,
			Status:      result.Status,
			Body:        result.Body,
			DurationMs:  result.Duration.Milliseconds(),
			ContentType: result.ContentType,
			Size:        result.Size,
			Request:     result.Request,
		}
		if result.Error != nil {
			cached.Error = result.Error.Error()
		}
		if result.Headers != nil {
			cached.Headers = make(map[string]string)
			for k, v := range result.Headers {
				if len(v) > 0 {
					cached.Headers[k] = v[0]
				}
			}
		}
		session.Responses[path] = cached
	}

	content, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		m.logger.Warnf("failed to marshal session cache: %v", err)
		return
	}

	err = os.WriteFile(m.sessionCacheFilePath(), content, 0644)
	if err != nil {
		m.logger.Warnf("failed to save session cache: %v", err)
	}
}

// loadCache loads the response cache and selected template path from disk
// Note: selected template is restored later after templates are loaded
func (m *Model) loadCache() {
	content, err := os.ReadFile(m.sessionCacheFilePath())
	if err != nil {
		// File doesn't exist or can't be read - that's fine
		return
	}

	var session SessionCache
	err = json.Unmarshal(content, &session)
	if err != nil {
		m.logger.Warnf("failed to parse session cache: %v", err)
		return
	}

	// Store the selected template path for later restoration
	m.pendingSelectedTemplate = session.SelectedTemplate

	m.responseCache = make(map[string]*RequestResult)
	for path, cached := range session.Responses {
		result := &RequestResult{
			StatusCode:  cached.StatusCode,
			Status:      cached.Status,
			Body:        cached.Body,
			Duration:    time.Duration(cached.DurationMs) * time.Millisecond,
			ContentType: cached.ContentType,
			Size:        cached.Size,
			Request:     cached.Request,
		}
		if cached.Error != "" {
			result.Error = fmt.Errorf("%s", cached.Error)
		}
		if cached.Headers != nil {
			result.Headers = make(map[string][]string)
			for k, v := range cached.Headers {
				result.Headers[k] = []string{v}
			}
		}
		m.responseCache[path] = result
	}
}

// openOverridesInEditor opens the overrides YAML file in the user's editor
func (m Model) openOverridesInEditor() tea.Cmd {
	filePath := m.overridesFilePath()

	// Create the file with a helpful header if it doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		header := `# Override variables for tp templates
# Format: key: value
# Example:
#   id: "12345"
#   name: "test"
#
# These values will override any matching variables in your templates.
`
		if err := os.WriteFile(filePath, []byte(header), 0644); err != nil {
			m.logger.Warnf("failed to create overrides file: %v", err)
		}
	}

	return m.openEditor(filePath)
}

func (m Model) handleNewTemplateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := m.newTemplateInput.Value()
		if name != "" {
			// Create the template file with default content
			templatePath, err := paths.NewAbsoluteFromRelative(name+".yml", m.config.TemplatesDirectoryPath)
			if err == nil {
				// Write default template content
				err = os.WriteFile(templatePath, static.DefaultTemplate, 0644)
				if err == nil {
					// Open in editor
					m.newTemplateInput.SetValue("")
					m.mode = ModeNormal
					m.newTemplateInput.Blur()
					return m, m.openEditor(templatePath)
				}
			}
		}
		m.newTemplateInput.SetValue("")
		m.mode = ModeNormal
		m.newTemplateInput.Blur()
		return m, nil
	case "esc":
		m.newTemplateInput.SetValue("")
		m.mode = ModeNormal
		m.newTemplateInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.newTemplateInput, cmd = m.newTemplateInput.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.pendingDeleteTemplate != nil {
			err := os.Remove(m.pendingDeleteTemplate.AbsolutePath)
			if err != nil {
				m.statusMsg = "Failed to delete: " + err.Error()
			} else {
				m.statusMsg = "Deleted: " + m.pendingDeleteTemplate.Name
				// Clear selection if it was the deleted template
				if m.selectedTemplate == m.pendingDeleteTemplate {
					m.selectedTemplate = nil
					m.response = nil
				}
				// Remove from cache
				delete(m.responseCache, m.pendingDeleteTemplate.AbsolutePath)
			}
			m.pendingDeleteTemplate = nil
			m.mode = ModeNormal
			return m, tea.Batch(m.loadTemplates, m.clearStatusAfterDelay())
		}
		m.mode = ModeNormal
		return m, nil
	case "n", "N", "esc":
		m.pendingDeleteTemplate = nil
		m.mode = ModeNormal
		return m, nil
	}
	return m, nil
}

func (m Model) openEditor(filePath string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	c := exec.Command(editor, filePath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err: err}
	})
}

func (m Model) clearStatusAfterDelay() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m Model) executeRequest() tea.Msg {
	if m.selectedTemplate == nil {
		return requestCompleteMsg{result: &RequestResult{Error: fmt.Errorf("no template selected")}}
	}

	result := ExecuteRequest(m.logger, m.selectedTemplate.AbsolutePath, m.vars, m.overrides)
	return requestCompleteMsg{result: result, templatePath: m.selectedTemplate.AbsolutePath}
}

func (m *Model) updateViewportSizes() {
	// Calculate panel sizes
	sidebarWidth := m.width / 4
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}
	if sidebarWidth > 50 {
		sidebarWidth = 50
	}

	mainWidth := m.width - sidebarWidth - 6 // Account for borders
	panelHeight := (m.height - 6) / 2       // Two panels vertically

	m.requestViewport.Width = mainWidth - 4
	m.requestViewport.Height = panelHeight - 6

	m.responseViewport.Width = mainWidth - 4
	m.responseViewport.Height = panelHeight - 6
}

// View renders the TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Calculate layout
	sidebarWidth := m.width / 4
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}
	if sidebarWidth > 50 {
		sidebarWidth = 50
	}

	mainWidth := m.width - sidebarWidth - 4
	panelHeight := (m.height - 5) / 2

	// Build panels
	leftPanel := m.renderTemplatesPanel(sidebarWidth-2, m.height-4)
	requestPanel := m.renderRequestPanel(mainWidth-2, panelHeight-1)
	responsePanel := m.renderResponsePanel(mainWidth-2, panelHeight-1)

	// Combine main panels
	rightSide := lipgloss.JoinVertical(lipgloss.Left, requestPanel, responsePanel)

	// Main content
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightSide)

	// Status bar
	statusBar := m.renderStatusBar()

	fullView := lipgloss.JoinVertical(lipgloss.Left, mainContent, statusBar)

	return fullView
}

func (m Model) renderTemplatesPanel(width, height int) string {
	style := getPanelStyle()
	if m.activePanel == PanelTemplates {
		style = getActivePanelStyle()
	}

	// Title
	title := titleStyle.Render("  Templates")

	// Search bar
	searchBar := ""
	if m.mode == ModeSearch {
		searchBar = focusedInputStyle.Width(width - 4).Render(m.searchInput.View())
	} else if m.searchInput.Value() != "" {
		searchBar = inputStyle.Width(width - 4).Render("/" + m.searchInput.Value())
	}

	// Template list
	listHeight := height - 4
	if searchBar != "" {
		listHeight -= 3
	}

	var items []string
	usedLines := 0

	// Calculate visible range - account for selected item taking 2 lines
	cursor := m.templates.Cursor()
	allItems := m.templates.Items()

	// Handle empty list
	if len(allItems) == 0 {
		emptyMsg := mutedStyle.Render("No templates found")
		items = append(items, emptyMsg)
		usedLines = 1
	} else {
		// Find start index that keeps cursor visible
		startIdx := 0

		// Adjust start if cursor would be off screen
		if cursor >= listHeight {
			startIdx = cursor - listHeight + 1
		}

		for i := startIdx; i < len(allItems) && usedLines < listHeight; i++ {
			item := allItems[i]
			if item == nil {
				continue
			}
			rendered := m.renderTemplateItem(item, i == cursor, width-6)
			items = append(items, rendered)
			usedLines++
		}
	}

	// Pad with empty lines if needed
	for usedLines < listHeight {
		items = append(items, strings.Repeat(" ", width-4))
		usedLines++
	}

	list := strings.Join(items, "\n")

	content := title + "\n"
	if searchBar != "" {
		content += searchBar + "\n"
	}
	content += list

	return style.Width(width).Height(height).Render(content)
}

func (m Model) renderTemplateItem(item *TemplateItem, selected bool, width int) string {
	if item == nil {
		return strings.Repeat(" ", width)
	}
	indent := strings.Repeat("  ", item.Depth)

	var lines []string
	if item.IsDir {
		arrow := "▸"
		if item.Expanded {
			arrow = "▾"
		}
		line := fmt.Sprintf("%s%s %s/", indent, arrow, item.Name)
		// Truncate if too long
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		line = fmt.Sprintf("%-*s", width, line)
		if selected {
			return selectedItemStyle.Render(line)
		}
		return normalItemStyle.Render(line)
	}

	// For template files, show method + name on first line
	methodStr := fmt.Sprintf("%-6s", item.Method)

	// Calculate visible width for truncation (without ANSI codes)
	// Layout: indent + method (6 chars) + " " + name
	prefixWidth := len(indent) + 6 + 1
	availableNameWidth := width - prefixWidth

	name := item.Name
	if availableNameWidth > 3 && len(name) > availableNameWidth {
		name = name[:availableNameWidth-3] + "..."
	} else if availableNameWidth <= 3 {
		name = ""
	}

	// Pad name to fill available space
	if len(name) < availableNameWidth {
		name = name + strings.Repeat(" ", availableNameWidth-len(name))
	}

	// Build styled line - apply method color, then wrap entire line in selection style if needed
	methodStyle := getMethodStyle(item.Method)
	styledMethod := methodStyle.Render(methodStr)

	if selected {
		lines = append(lines, fmt.Sprintf("%s%s %s",
			selectedItemStyle.Render(indent),
			methodStyle.Copy().Background(highlightBg).Render(methodStr),
			selectedItemStyle.Render(name)))
	} else {
		lines = append(lines, fmt.Sprintf("%s%s %s", indent, styledMethod, normalItemStyle.Render(name)))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderRequestPanel(width, height int) string {
	style := getPanelStyle()
	if m.activePanel == PanelRequest {
		style = getActivePanelStyle()
	}

	// Title with tabs
	headerTab := inactiveTabStyle.Render("Headers [h]")
	bodyTab := inactiveTabStyle.Render("Body [b]")
	paramsTab := inactiveTabStyle.Render("Params [p]")

	switch m.requestTab {
	case TabHeaders:
		headerTab = activeTabStyle.Render("Headers [h]")
	case TabBody:
		bodyTab = activeTabStyle.Render("Body [b]")
	case TabParams:
		paramsTab = activeTabStyle.Render("Params [p]")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, headerTab, " ", bodyTab, " ", paramsTab)
	title := titleStyle.Render("  Request") + "  " + tabs

	// Method and URL
	var methodURL string
	if m.selectedTemplate != nil {
		methodStyle := getMethodStyle(m.selectedTemplate.Method)
		method := methodStyle.Render(m.selectedTemplate.Method)
		url := urlStyle.Render(truncate(m.selectedTemplate.URL, width-15))
		methodURL = method + " " + url
	} else {
		methodURL = mutedStyle.Render("Select a template to view request details")
	}

	content := title + "\n" + methodURL + "\n\n"

	// Viewport content
	if m.selectedTemplate != nil {
		content += m.requestViewport.View()
	}

	return style.Width(width).Height(height).Render(content)
}

func (m Model) renderRequestDetails() string {
	if m.selectedTemplate == nil {
		return ""
	}

	var content strings.Builder

	switch m.requestTab {
	case TabHeaders:
		if len(m.selectedTemplate.Headers) == 0 {
			content.WriteString(mutedStyle.Render("No headers defined"))
		} else {
			for k, v := range m.selectedTemplate.Headers {
				content.WriteString(headerKeyStyle.Render(k))
				content.WriteString(": ")
				content.WriteString(headerValueStyle.Render(v))
				content.WriteString("\n")
			}
		}
	case TabBody:
		if m.selectedTemplate.Body == "" {
			content.WriteString(mutedStyle.Render("No body defined"))
		} else {
			// Try to detect content type from headers for syntax highlighting
			lexer := ""
			if ct, ok := m.selectedTemplate.Headers["Content-Type"]; ok {
				lexer = getLexerForContentType(ct)
			}
			if lexer != "" {
				// Use template-aware highlighting since request bodies may contain Go template tags
				content.WriteString(highlightCodeWithTemplates(m.selectedTemplate.Body, lexer))
			} else {
				content.WriteString(m.selectedTemplate.Body)
			}
		}
	case TabParams:
		// Show template variables/descriptions
		if len(m.selectedTemplate.Descriptions) == 0 {
			content.WriteString(mutedStyle.Render("No parameters documented"))
		} else {
			for k, v := range m.selectedTemplate.Descriptions {
				content.WriteString(overrideKeyStyle.Render(k))
				content.WriteString(": ")
				content.WriteString(subtitleStyle.Render(v))
				content.WriteString("\n")
			}
		}
	}

	return content.String()
}

func (m Model) renderResponsePanel(width, height int) string {
	style := getPanelStyle()
	if m.activePanel == PanelResponse {
		style = getActivePanelStyle()
	}

	// Title with tabs
	headerTab := inactiveTabStyle.Render("Headers [h]")
	bodyTab := inactiveTabStyle.Render("Body [b]")

	switch m.responseTab {
	case TabHeaders:
		headerTab = activeTabStyle.Render("Headers [h]")
	case TabBody:
		bodyTab = activeTabStyle.Render("Body [b]")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Left, headerTab, " ", bodyTab)
	title := titleStyle.Render("  Response") + "  " + tabs

	// Status line
	var statusLine string
	if m.loading {
		statusLine = m.spinner.View() + " Sending request..."
	} else if m.response != nil {
		if m.response.Error != nil {
			statusLine = errorStatusStyle.Render("Error: " + m.response.Error.Error())
		} else {
			status := renderStatusCode(m.response.StatusCode)
			duration := fmt.Sprintf("%.2fms", float64(m.response.Duration.Microseconds())/1000)
			size := fmt.Sprintf("%d bytes", m.response.Size)
			statusLine = fmt.Sprintf("%s  %s  %s", status, mutedStyle.Render(duration), mutedStyle.Render(size))
		}
	} else {
		statusLine = mutedStyle.Render("Press 'x' to send request")
	}

	content := title + "\n" + statusLine + "\n\n"
	content += m.responseViewport.View()

	return style.Width(width).Height(height).Render(content)
}

func (m Model) renderResponseBody() string {
	if m.response == nil {
		return ""
	}

	if m.response.Error != nil {
		return errorStatusStyle.Render(m.response.Error.Error())
	}

	switch m.responseTab {
	case TabHeaders:
		var content strings.Builder
		for k, v := range m.response.Headers {
			content.WriteString(headerKeyStyle.Render(k))
			content.WriteString(": ")
			content.WriteString(headerValueStyle.Render(strings.Join(v, ", ")))
			content.WriteString("\n")
		}
		return content.String()
	case TabBody:
		lexer := getLexerForContentType(m.response.ContentType)
		if lexer != "" {
			return highlightCode(m.response.Body, lexer)
		}
		return m.response.Body
	}

	return ""
}

func (m Model) renderStatusBar() string {
	// Left side - mode/status
	var leftContent string
	switch m.mode {
	case ModeSearch:
		leftContent = " SEARCH "
	case ModeNewTemplate:
		leftContent = " NEW: " + m.newTemplateInput.View() + " "
	case ModeConfirmDelete:
		leftContent = fmt.Sprintf(" DELETE '%s'? [y/n] ", m.pendingDeleteTemplate.Name)
	default:
		leftContent = " NORMAL "
	}

	// Middle - status message or overrides count
	middleContent := ""
	if m.statusMsg != "" {
		middleContent = " " + m.statusMsg + " "
	} else if len(m.overrides) > 0 {
		middleContent = fmt.Sprintf(" | Overrides: %d ", len(m.overrides))
		for _, o := range m.overrides {
			middleContent += fmt.Sprintf("[%s=%s] ", o.Key, truncate(o.Value, 10))
		}
	}

	// Right side - help
	help := " q:quit  /:search  x:send  y:copy  o:override  e:edit  n:new  d:delete  ?:help "
	if !m.showHelp {
		help = " ?:help "
	}

	// Calculate widths
	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(help)
	middleWidth := m.width - leftWidth - rightWidth

	if middleWidth < 0 {
		middleWidth = 0
	}

	middle := fmt.Sprintf("%-*s", middleWidth, middleContent)

	return statusBarStyle.Render(leftContent + middle + help)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Run starts the TUI
func Run(logger *logging.Logger, cfg config.Config, envFile string) error {
	p := tea.NewProgram(
		New(logger, cfg, envFile),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	return err
}

package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dfairburn/tp/config"
	"github.com/dfairburn/tp/handlers"
	"github.com/dfairburn/tp/paths"
	"github.com/dfairburn/tp/static"
	logging "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// TemplateItem represents a request template in the list
type TemplateItem struct {
	Name         string
	AbsolutePath string
	RelativePath string
	Method       string
	URL          string
	Headers      map[string]string
	Body         string
	Descriptions map[string]string
	IsDir        bool
	Depth        int
	Expanded     bool
	Children     []*TemplateItem
	Parent       *TemplateItem
}

// TemplateList manages the list of templates
type TemplateList struct {
	items         []*TemplateItem
	flatItems     []*TemplateItem // Flattened view for display
	cursor        int
	templatesDir  string
	selectedItem  *TemplateItem
	searchQuery   string
	filteredItems []*TemplateItem
}

// NewTemplateList creates a new template list
func NewTemplateList(templatesDir string) *TemplateList {
	tl := &TemplateList{
		templatesDir: paths.Expand(templatesDir),
		items:        make([]*TemplateItem, 0),
		flatItems:    make([]*TemplateItem, 0),
	}
	return tl
}

// Load reads all templates from the templates directory
func (tl *TemplateList) Load(logger *logging.Logger) error {
	tl.items = make([]*TemplateItem, 0)

	re, err := regexp.Compile(static.YamlRegex)
	if err != nil {
		return err
	}

	// Build a tree structure
	dirMap := make(map[string]*TemplateItem)

	walkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip the root directory itself
		if p == tl.templatesDir {
			return nil
		}

		relPath, _ := filepath.Rel(tl.templatesDir, p)
		depth := strings.Count(relPath, string(os.PathSeparator))

		if info.IsDir() {
			item := &TemplateItem{
				Name:         info.Name(),
				AbsolutePath: p,
				RelativePath: relPath,
				IsDir:        true,
				Depth:        depth,
				Expanded:     true,
				Children:     make([]*TemplateItem, 0),
			}
			dirMap[p] = item

			// Find parent
			parentPath := filepath.Dir(p)
			if parent, ok := dirMap[parentPath]; ok {
				parent.Children = append(parent.Children, item)
				item.Parent = parent
			} else {
				tl.items = append(tl.items, item)
			}
			return nil
		}

		// Check if it's a yaml file
		if !re.MatchString(p) {
			return nil
		}

		// Load template details
		item := &TemplateItem{
			Name:         strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
			AbsolutePath: p,
			RelativePath: relPath,
			IsDir:        false,
			Depth:        depth,
		}

		// Try to load template metadata
		if err := item.LoadMetadata(); err != nil {
			// Default to GET if we can't load
			item.Method = "GET"
		}

		// Find parent directory
		parentPath := filepath.Dir(p)
		if parent, ok := dirMap[parentPath]; ok {
			parent.Children = append(parent.Children, item)
			item.Parent = parent
		} else {
			tl.items = append(tl.items, item)
		}

		return nil
	}

	// Use config.LoadTemplateFiles which handles hidden file skipping
	err = config.LoadTemplateFiles(logger, tl.templatesDir, walkFunc)
	if err != nil {
		return err
	}

	// Sort items
	tl.sortItems(tl.items)

	// Flatten for display
	tl.flatten()

	return nil
}

func (tl *TemplateList) sortItems(items []*TemplateItem) {
	sort.Slice(items, func(i, j int) bool {
		// Directories first
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	for _, item := range items {
		if item.IsDir && len(item.Children) > 0 {
			tl.sortItems(item.Children)
		}
	}
}

func (tl *TemplateList) flatten() {
	tl.flatItems = make([]*TemplateItem, 0)
	tl.flattenRecursive(tl.items)
	tl.filteredItems = tl.flatItems
}

func (tl *TemplateList) flattenRecursive(items []*TemplateItem) {
	for _, item := range items {
		tl.flatItems = append(tl.flatItems, item)
		if item.IsDir && item.Expanded && len(item.Children) > 0 {
			tl.flattenRecursive(item.Children)
		}
	}
}

// LoadMetadata loads the template's metadata from its YAML file
func (t *TemplateItem) LoadMetadata() error {
	content, err := os.ReadFile(t.AbsolutePath)
	if err != nil {
		return err
	}

	// Use the handlers.Template struct to parse the YAML
	var tmpl handlers.Template
	err = yaml.Unmarshal(content, &tmpl)
	if err != nil {
		return err
	}

	t.Method = strings.ToUpper(tmpl.Method)
	if t.Method == "" {
		t.Method = "GET"
	}
	t.URL = tmpl.Url
	t.Headers = tmpl.Headers
	t.Body = tmpl.Body
	t.Descriptions = tmpl.Descriptions

	return nil
}

// Filter filters the items based on a search query.
// When a child matches, its parent folders are also included for context.
func (tl *TemplateList) Filter(query string) {
	tl.searchQuery = strings.ToLower(query)

	if tl.searchQuery == "" {
		tl.filteredItems = tl.flatItems
		return
	}

	// First pass: find all matching items and collect their parents
	matchingItems := make(map[*TemplateItem]bool)
	for _, item := range tl.flatItems {
		if strings.Contains(strings.ToLower(item.Name), tl.searchQuery) {
			matchingItems[item] = true
			// Also include all parent folders for context
			parent := item.Parent
			for parent != nil {
				matchingItems[parent] = true
				parent = parent.Parent
			}
		}
	}

	// Second pass: build filtered list preserving order from flatItems
	tl.filteredItems = make([]*TemplateItem, 0)
	for _, item := range tl.flatItems {
		if matchingItems[item] {
			tl.filteredItems = append(tl.filteredItems, item)
		}
	}

	// Reset cursor if out of bounds
	if tl.cursor >= len(tl.filteredItems) {
		tl.cursor = max(0, len(tl.filteredItems)-1)
	}
}

// Navigation methods
func (tl *TemplateList) MoveUp() {
	if tl.cursor > 0 {
		tl.cursor--
	}
}

func (tl *TemplateList) MoveDown() {
	if tl.cursor < len(tl.filteredItems)-1 {
		tl.cursor++
	}
}

func (tl *TemplateList) PageUp(pageSize int) {
	tl.cursor -= pageSize
	if tl.cursor < 0 {
		tl.cursor = 0
	}
}

func (tl *TemplateList) PageDown(pageSize int) {
	tl.cursor += pageSize
	if tl.cursor >= len(tl.filteredItems) {
		tl.cursor = max(0, len(tl.filteredItems)-1)
	}
}

func (tl *TemplateList) ToggleExpand() {
	if len(tl.filteredItems) == 0 {
		return
	}
	item := tl.filteredItems[tl.cursor]
	if item.IsDir {
		item.Expanded = !item.Expanded
		tl.flatten()
		tl.Filter(tl.searchQuery)
	}
}

func (tl *TemplateList) Select() *TemplateItem {
	if len(tl.filteredItems) == 0 {
		return nil
	}
	item := tl.filteredItems[tl.cursor]
	if !item.IsDir {
		tl.selectedItem = item
		return item
	}
	// Toggle expansion for directories
	tl.ToggleExpand()
	return nil
}

func (tl *TemplateList) Selected() *TemplateItem {
	return tl.selectedItem
}

func (tl *TemplateList) Current() *TemplateItem {
	if len(tl.filteredItems) == 0 {
		return nil
	}
	return tl.filteredItems[tl.cursor]
}

func (tl *TemplateList) Cursor() int {
	return tl.cursor
}

func (tl *TemplateList) Items() []*TemplateItem {
	return tl.filteredItems
}

func (tl *TemplateList) Len() int {
	return len(tl.filteredItems)
}

// ClearSearchAndFocus clears the search filter and positions the cursor on the given item.
// If the item is inside a collapsed folder, it expands the parent folders.
func (tl *TemplateList) ClearSearchAndFocus(item *TemplateItem) {
	tl.searchQuery = ""

	// Ensure all parent folders are expanded so the item is visible
	if item != nil {
		parent := item.Parent
		for parent != nil {
			parent.Expanded = true
			parent = parent.Parent
		}
	}

	// Re-flatten with all necessary folders expanded
	tl.flatten()

	// Find the item in the flattened list and set cursor
	if item != nil {
		for i, it := range tl.flatItems {
			if it == item {
				tl.cursor = i
				break
			}
		}
	}
}

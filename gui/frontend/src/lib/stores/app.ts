import { writable, derived, get } from 'svelte/store';
import type { main } from '../../wailsjs/go/models';

// Storage keys
const STORAGE_KEYS = {
  selectedTemplatePath: 'tp-gui-selected-template',
  expandedDirs: 'tp-gui-expanded-dirs',
  responseCache: 'tp-gui-response-cache',
  theme: 'tp-gui-theme',
};

// Helper to safely parse JSON from localStorage
function loadFromStorage<T>(key: string, defaultValue: T): T {
  try {
    const stored = localStorage.getItem(key);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (e) {
    console.warn(`Failed to load ${key} from storage:`, e);
  }
  return defaultValue;
}

// Helper to save to localStorage
function saveToStorage(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch (e) {
    console.warn(`Failed to save ${key} to storage:`, e);
  }
}

// Template list state
export const templates = writable<main.TemplateItem[]>([]);
export const selectedTemplate = writable<main.TemplateItem | null>(null);
export const selectedTemplatePath = writable<string>(loadFromStorage(STORAGE_KEYS.selectedTemplatePath, ''));
export const expandedDirs = writable<Set<string>>(new Set(loadFromStorage<string[]>(STORAGE_KEYS.expandedDirs, [])));
export const searchQuery = writable<string>('');

// Response state
export const currentResponse = writable<main.HTTPResponse | null>(null);
export const isExecuting = writable<boolean>(false);

// Response cache: Map of template path -> response
export const responseCache = writable<Record<string, main.HTTPResponse>>(
  loadFromStorage(STORAGE_KEYS.responseCache, {})
);

// UI state
export const activePanel = writable<'templates' | 'request' | 'response'>('templates');
export const requestTab = writable<'headers' | 'body' | 'params'>('body');
export const responseTab = writable<'body' | 'headers'>('body');
export const theme = writable<'dark' | 'light'>(loadFromStorage(STORAGE_KEYS.theme, 'dark'));

// Variables/overrides for the current request
export const overrides = writable<Record<string, string>>({});

// Template editor modal state
export const showTemplateEditor = writable<boolean>(false);
export const editorTemplateItem = writable<main.TemplateItem | null>(null);

// Persist selected template path
selectedTemplatePath.subscribe(value => {
  saveToStorage(STORAGE_KEYS.selectedTemplatePath, value);
});

// Persist expanded dirs
expandedDirs.subscribe(value => {
  saveToStorage(STORAGE_KEYS.expandedDirs, Array.from(value));
});

// Persist theme
theme.subscribe(value => {
  saveToStorage(STORAGE_KEYS.theme, value);
});

// Persist response cache (limit size to prevent storage overflow)
responseCache.subscribe(value => {
  // Keep only the last 50 responses
  const entries = Object.entries(value);
  if (entries.length > 50) {
    const trimmed = Object.fromEntries(entries.slice(-50));
    saveToStorage(STORAGE_KEYS.responseCache, trimmed);
  } else {
    saveToStorage(STORAGE_KEYS.responseCache, value);
  }
});

// When selected template changes, update path and load cached response
selectedTemplate.subscribe(template => {
  if (template && !template.isDir) {
    selectedTemplatePath.set(template.absolutePath);
    // Load cached response if available
    const cache = get(responseCache);
    if (cache[template.absolutePath]) {
      currentResponse.set(cache[template.absolutePath]);
    } else {
      currentResponse.set(null);
    }
  }
});

// When response changes, cache it
currentResponse.subscribe(response => {
  const template = get(selectedTemplate);
  if (response && template && !response.error) {
    responseCache.update(cache => ({
      ...cache,
      [template.absolutePath]: response,
    }));
  }
});

// Helper to restore selected template from templates list
export function restoreSelectedTemplate(items: main.TemplateItem[]): main.TemplateItem | null {
  const savedPath = get(selectedTemplatePath);
  if (!savedPath) return null;
  
  function findByPath(items: main.TemplateItem[]): main.TemplateItem | null {
    for (const item of items) {
      if (item.absolutePath === savedPath) {
        return item;
      }
      if (item.children) {
        const found = findByPath(item.children);
        if (found) return found;
      }
    }
    return null;
  }
  
  return findByPath(items);
}

// Filtered templates based on search
export const filteredTemplates = derived(
  [templates, searchQuery, expandedDirs],
  ([$templates, $searchQuery, $expandedDirs]) => {
    const query = $searchQuery.toLowerCase();
    
    function flattenAndFilter(items: main.TemplateItem[], depth: number = 0): main.TemplateItem[] {
      const result: main.TemplateItem[] = [];
      
      for (const item of items) {
        const matches = !query || item.name.toLowerCase().includes(query);
        const hasMatchingChildren = item.children?.some(child => 
          child.name.toLowerCase().includes(query) || 
          (child.isDir && hasMatchingDescendants(child, query))
        );
        
        if (matches || hasMatchingChildren) {
          result.push(item);
          
          if (item.isDir && item.children && ($expandedDirs.has(item.absolutePath) || query)) {
            result.push(...flattenAndFilter(item.children, depth + 1));
          }
        }
      }
      
      return result;
    }
    
    function hasMatchingDescendants(item: main.TemplateItem, q: string): boolean {
      if (!item.children) return false;
      return item.children.some(child => 
        child.name.toLowerCase().includes(q) || 
        (child.isDir && hasMatchingDescendants(child, q))
      );
    }
    
    return flattenAndFilter($templates);
  }
);

// Helper to toggle directory expansion
export function toggleDirectory(path: string) {
  expandedDirs.update(dirs => {
    const newDirs = new Set(dirs);
    if (newDirs.has(path)) {
      newDirs.delete(path);
    } else {
      newDirs.add(path);
    }
    return newDirs;
  });
}

// Helper to open template editor
export function openTemplateEditor(item: main.TemplateItem | null = null) {
  editorTemplateItem.set(item);
  showTemplateEditor.set(true);
}

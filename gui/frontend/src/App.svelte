<script lang="ts">
  import TemplateList from './lib/components/TemplateList.svelte';
  import RequestPanel from './lib/components/RequestPanel.svelte';
  import ResponsePanel from './lib/components/ResponsePanel.svelte';
  import SettingsPanel from './lib/components/SettingsPanel.svelte';
  import { theme, selectedTemplate, isExecuting, currentResponse } from './lib/stores/app';
  import { ExecuteTemplateWithOverrides } from '../wailsjs/go/app/App';
  import { makeErrorResponse } from './lib/utils';
  import { onMount, onDestroy } from 'svelte';

  let settingsOpen = false;

  // Vertical split between request and response panels
  let vertSplitPercent = 50;
  let vertIsDragging = false;
  let panelsContainer: HTMLElement | null = null;

  function startVertDrag(e: MouseEvent) {
    vertIsDragging = true;
    e.preventDefault();
    window.addEventListener('mousemove', onVertDragMove);
    window.addEventListener('mouseup', stopVertDrag);
  }

  function onVertDragMove(e: MouseEvent) {
    if (!panelsContainer) return;
    const rect = panelsContainer.getBoundingClientRect();
    const y = e.clientY - rect.top;
    vertSplitPercent = Math.max(15, Math.min(85, (y / rect.height) * 100));
  }

  function stopVertDrag() {
    vertIsDragging = false;
    window.removeEventListener('mousemove', onVertDragMove);
    window.removeEventListener('mouseup', stopVertDrag);
  }

  onDestroy(() => {
    window.removeEventListener('mousemove', onVertDragMove);
    window.removeEventListener('mouseup', stopVertDrag);
  });

  // Handle keyboard shortcuts
  function handleKeydown(e: KeyboardEvent) {
    // Don't handle shortcuts when settings is open
    if (settingsOpen && e.key !== 'Escape') return;

    // Cmd/Ctrl + Enter to execute
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      executeCurrentTemplate();
    }
    
    // Cmd/Ctrl + , to open settings
    if ((e.metaKey || e.ctrlKey) && e.key === ',') {
      e.preventDefault();
      settingsOpen = true;
    }
    
    // Escape to close settings or clear selection
    if (e.key === 'Escape') {
      if (settingsOpen) {
        settingsOpen = false;
      } else {
        selectedTemplate.set(null);
        currentResponse.set(null);
      }
    }
  }

  async function executeCurrentTemplate() {
    const template = $selectedTemplate;
    if (!template || template.isDir) return;
    
    isExecuting.set(true);
    currentResponse.set(null);
    
    try {
      // Use ExecuteTemplateWithOverrides to include file-based overrides
      const response = await ExecuteTemplateWithOverrides(template.absolutePath, {});
      currentResponse.set(response);
    } catch (err) {
      currentResponse.set(makeErrorResponse(err));
    } finally {
      isExecuting.set(false);
    }
  }

  onMount(() => {
    // Only detect system theme on first visit (when no stored preference)
    const storedTheme = localStorage.getItem('tp-gui-theme');
    if (!storedTheme) {
      if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
        theme.set('light');
      }
    }
  });
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="app" class:light={$theme === 'light'}>
  <div class="sidebar">
    <TemplateList />
  </div>
  <div class="main">
    <div class="toolbar">
      <div class="toolbar-left">
      </div>
      <div class="toolbar-actions">
        <button
          class="toolbar-btn theme-toggle-btn"
          on:click={() => theme.update(t => t === 'dark' ? 'light' : 'dark')}
          title="Toggle theme"
        >
          {$theme === 'dark' ? '☀' : '🌙'}
        </button>
        <button
          class="toolbar-btn"
          on:click={() => settingsOpen = true}
          title="Settings (⌘,)"
        >
          ⚙ Settings
        </button>
      </div>
    </div>
    <div class="panels" bind:this={panelsContainer}>
      <div class="request-area" style="flex: 0 0 {vertSplitPercent}%">
        <RequestPanel />
      </div>
      <div
        class="vert-resize-handle"
        class:dragging={vertIsDragging}
        on:mousedown={startVertDrag}
        role="separator"
        aria-label="Resize request/response panels"
      ></div>
      <div class="response-area">
        <ResponsePanel />
      </div>
    </div>
  </div>
  
  <SettingsPanel bind:isOpen={settingsOpen} />
</div>

<style>
  :global(:root) {
    /* Dark theme (default) */
    --bg-primary: #1e1e1e;
    --bg-secondary: #252526;
    --bg-hover: #2a2d2e;
    --bg-selected: #37373d;
    --text-primary: #cccccc;
    --text-secondary: #9d9d9d;
    --text-muted: #6e6e6e;
    --border-color: #3c3c3c;
    --accent-color: #0078d4;
  }

  :global(.light) {
    --bg-primary: #ffffff;
    --bg-secondary: #f3f3f3;
    --bg-hover: #e8e8e8;
    --bg-selected: #d4d4d4;
    --text-primary: #1e1e1e;
    --text-secondary: #656565;
    --text-muted: #9e9e9e;
    --border-color: #e0e0e0;
    --accent-color: #0078d4;
  }

  :global(*) {
    box-sizing: border-box;
  }

  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .app {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .sidebar {
    width: 280px;
    min-width: 200px;
    max-width: 400px;
    flex-shrink: 0;
  }

  .main {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }

  .panels {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
    gap: 12px;
  }

  .toolbar-left {
    flex: 1;
  }

  .toolbar-actions {
    display: flex;
    gap: 8px;
  }

  .toolbar-btn {
    padding: 6px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .toolbar-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .theme-toggle-btn {
    font-size: 14px;
    padding: 4px 8px;
  }

  .request-area {
    min-height: 0;
    overflow: hidden;
  }

  .vert-resize-handle {
    flex: 0 0 5px;
    cursor: row-resize;
    background: var(--border-color);
    transition: background 0.15s;
    user-select: none;
  }

  .vert-resize-handle:hover,
  .vert-resize-handle.dragging {
    background: var(--accent-color);
  }

  .response-area {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }
</style>

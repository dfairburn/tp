<script lang="ts">
  import { onMount, onDestroy, afterUpdate } from 'svelte';
  import { currentResponse, responseTab, isExecuting } from '../stores/app';
  import {
    getStatusColor, formatHeaders, formatDuration, formatSize,
    getResponseSize, getHighlightedBody, applySearchHighlights,
  } from './ResponsePanel';

  $: response = $currentResponse;

  // ── Search state ──────────────────────────────────────────────────────────
  let searchVisible = false;
  let searchQuery = '';
  let matchIndex = 0;
  let matchCount = 0;
  let searchInput: HTMLInputElement;
  let preElement: HTMLElement;
  let headersPreElement: HTMLElement;

  // Derived display body — recomputes whenever the query, body, or visibility changes.
  // Highlights are debounced so a large body isn't re-scanned on every keystroke.
  let displayBody = '';
  let _searchTimer: ReturnType<typeof setTimeout> | null = null;
  $: {
    if (_searchTimer) clearTimeout(_searchTimer);
    const query = searchQuery;
    const body = highlightedBody;
    const visible = searchVisible;
    if (visible && query.trim()) {
      _searchTimer = setTimeout(() => {
        const result = applySearchHighlights(body, query);
        displayBody = result.html;
        matchCount = result.count;
      }, 150);
    } else {
      displayBody = body;
      matchCount = 0;
    }
  }

  // Reset to first match when the query changes
  $: if (searchQuery !== undefined) matchIndex = 0;

  // Use a plain object (not a reactive let) as a dirty flag so afterUpdate can
  // short-circuit without triggering extra re-renders.
  const _searchSync = { pending: false };
  $: if (searchVisible) { matchIndex; displayBody; _searchSync.pending = true; }

  // After DOM update, sync the current-match highlight class and scroll — only
  // when search state actually changed (guarded by _searchSync.pending).
  afterUpdate(() => {
    if (!_searchSync.pending || !searchVisible || !preElement) return;
    _searchSync.pending = false;
    const marks = Array.from(preElement.querySelectorAll<HTMLElement>('.search-mark'));
    if (!marks.length) return;
    const idx = Math.max(0, Math.min(matchIndex, marks.length - 1));
    marks.forEach((el, i) => el.classList.toggle('search-mark-current', i === idx));
    marks[idx]?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
  });

  function openSearch() {
    searchVisible = true;
    // Use setTimeout so the input is in the DOM before we focus it
    setTimeout(() => searchInput?.focus(), 0);
  }

  function closeSearch() {
    searchVisible = false;
    searchQuery = '';
    matchIndex = 0;
  }

  function nextMatch() {
    if (matchCount === 0) return;
    matchIndex = (matchIndex + 1) % matchCount;
  }

  function prevMatch() {
    if (matchCount === 0) return;
    matchIndex = (matchIndex - 1 + matchCount) % matchCount;
  }

  function handleSearchKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') { closeSearch(); return; }
    if (e.key === 'Enter') {
      e.preventDefault();
      e.shiftKey ? prevMatch() : nextMatch();
    }
  }

  function handleGlobalKeydown(e: KeyboardEvent) {
    const mod = e.metaKey || e.ctrlKey;
    if (mod && e.key === 'f') {
      e.preventDefault();
      openSearch();
      return;
    }
    if (mod && e.key === 'g') {
      e.preventDefault();
      e.shiftKey ? prevMatch() : nextMatch();
      return;
    }
    if (mod && e.key === 'a') {
      const tag = (document.activeElement as HTMLElement)?.tagName?.toLowerCase();
      if (tag !== 'input' && tag !== 'textarea') {
        const target = $responseTab === 'body' ? preElement : headersPreElement;
        if (target) {
          e.preventDefault();
          const sel = window.getSelection();
          const range = document.createRange();
          range.selectNodeContents(target);
          sel?.removeAllRanges();
          sel?.addRange(range);
        }
      }
      return;
    }
    if (e.key === 'Escape' && searchVisible) {
      closeSearch();
    }
  }

  onMount(() => { document.addEventListener('keydown', handleGlobalKeydown); });
  onDestroy(() => { document.removeEventListener('keydown', handleGlobalKeydown); });

  async function copyToClipboard() {
    if (!response?.body) return;
    try {
      await navigator.clipboard.writeText(response.body);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  }

  $: responseSize = response?.body ? getResponseSize(response.body) : 0;
  $: highlightedBody = response ? getHighlightedBody(response.body, response.contentType) : '';
</script>

<div class="response-panel">
  <div class="panel-header">
    <span class="title">Response</span>
    <div class="tabs">
      <button
        class="tab"
        class:active={$responseTab === 'body'}
        on:click={() => responseTab.set('body')}
      >
        Body
      </button>
      <button
        class="tab"
        class:active={$responseTab === 'headers'}
        on:click={() => responseTab.set('headers')}
      >
        Headers
      </button>
    </div>
  </div>

  {#if $isExecuting}
    <div class="loading">
      <div class="spinner"></div>
      <p>Executing request...</p>
    </div>
  {:else if response}
    {#if response.error}
      <div class="error">
        <span class="error-icon">⚠</span>
        <span class="error-text">{response.error}</span>
      </div>
    {:else}
      <div class="status-bar">
        <span class="status" style="color: {getStatusColor(response.statusCode)}">
          {response.statusCode} {response.status.replace(String(response.statusCode), '').trim()}
        </span>
        <span class="duration">{formatDuration(response.duration)}</span>
        <span class="size">{formatSize(responseSize)}</span>
        <span class="content-type">{response.contentType || 'Unknown'}</span>
        <button class="copy-btn" on:click={copyToClipboard} title="Copy response body">
          Copy
        </button>
      </div>

      {#if searchVisible && $responseTab === 'body'}
        <div class="search-bar">
          <input
            bind:this={searchInput}
            bind:value={searchQuery}
            on:keydown={handleSearchKeydown}
            class="search-input"
            placeholder="Search…"
            autocomplete="off"
            spellcheck="false"
          />
          <span class="match-count">
            {#if searchQuery.trim() && matchCount > 0}
              {matchIndex + 1} / {matchCount}
            {:else if searchQuery.trim()}
              No results
            {/if}
          </span>
          <button class="search-nav-btn" on:click={prevMatch} disabled={matchCount === 0} title="Previous (Shift+Enter)">↑</button>
          <button class="search-nav-btn" on:click={nextMatch} disabled={matchCount === 0} title="Next (Enter)">↓</button>
          <button class="search-close-btn" on:click={closeSearch} title="Close (Esc)">×</button>
        </div>
      {/if}

      <div class="content">
        {#if $responseTab === 'body'}
          <pre class="code-block highlighted" bind:this={preElement}>{@html displayBody || 'Empty response'}</pre>
        {:else if $responseTab === 'headers'}
          <pre class="code-block" bind:this={headersPreElement}>{formatHeaders(response.headers) || 'No headers'}</pre>
        {/if}
      </div>
    {/if}
  {:else}
    <div class="empty">
      <p>No response yet</p>
      <p class="hint">Execute a request to see the response</p>
    </div>
  {/if}
</div>

<style>
  .response-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-primary);
    border-top: 1px solid var(--border-color);
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .title {
    font-weight: 600;
    font-size: 13px;
    color: var(--text-primary);
  }

  .tabs {
    display: flex;
    gap: 4px;
  }

  .tab {
    padding: 4px 12px;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.15s;
  }

  .tab:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .tab.active {
    background: var(--accent-color);
    color: white;
  }

  .status-bar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .status {
    font-weight: 600;
    font-size: 13px;
  }

  .duration, .content-type, .size {
    font-size: 12px;
    color: var(--text-muted);
  }

  .size {
    padding: 2px 6px;
    background: var(--bg-hover);
    border-radius: 3px;
  }

  .copy-btn {
    margin-left: auto;
    padding: 4px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .copy-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  /* Search bar */
  .search-bar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .search-input {
    flex: 1;
    max-width: 280px;
    padding: 4px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    outline: none;
  }

  .search-input:focus {
    border-color: var(--accent-color);
  }

  .match-count {
    font-size: 11px;
    color: var(--text-muted);
    white-space: nowrap;
    min-width: 56px;
  }

  .search-nav-btn {
    padding: 3px 7px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.15s;
    line-height: 1;
  }

  .search-nav-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .search-nav-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .search-close-btn {
    padding: 3px 7px;
    border: none;
    border-radius: 4px;
    background: transparent;
    color: var(--text-muted);
    font-size: 15px;
    cursor: pointer;
    line-height: 1;
    transition: all 0.15s;
  }

  .search-close-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  /* Search match highlights (global because they're injected via {@html}) */
  :global(.search-mark) {
    background: rgba(255, 213, 0, 0.35);
    border-radius: 2px;
    color: inherit;
  }

  :global(.search-mark-current) {
    background: rgba(255, 140, 0, 0.6);
    border-radius: 2px;
    outline: 1px solid rgba(255, 140, 0, 0.9);
  }

  /* Light theme adjustments for marks */
  :global(.light) :global(.search-mark) {
    background: rgba(255, 213, 0, 0.5);
  }

  :global(.light) :global(.search-mark-current) {
    background: rgba(255, 140, 0, 0.5);
    outline-color: rgba(200, 100, 0, 0.8);
  }

  .content {
    flex: 1;
    overflow: auto;
    padding: 12px;
  }

  .code-block {
    margin: 0;
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    line-height: 1.5;
    color: var(--text-primary);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
  }

  .spinner {
    width: 24px;
    height: 24px;
    border: 2px solid var(--border-color);
    border-top-color: var(--accent-color);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .loading p {
    margin-top: 12px;
  }

  .error {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 16px;
    margin: 12px;
    background: rgba(249, 62, 62, 0.1);
    border: 1px solid rgba(249, 62, 62, 0.3);
    border-radius: 8px;
    color: #f93e3e;
  }

  .error-icon {
    font-size: 16px;
  }

  .error-text {
    font-family: 'SF Mono', Monaco, monospace;
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
    text-align: center;
  }

  .empty p {
    margin: 4px 0;
  }

  .hint {
    font-size: 12px;
    opacity: 0.7;
  }

  /* highlight.js syntax highlighting - Dark theme (VS Code inspired) */
  .code-block.highlighted :global(.hljs-string),
  .code-block.highlighted :global(.hljs-attr) {
    color: #ce9178;
  }

  .code-block.highlighted :global(.hljs-number) {
    color: #b5cea8;
  }

  .code-block.highlighted :global(.hljs-literal),
  .code-block.highlighted :global(.hljs-keyword) {
    color: #569cd6;
  }

  .code-block.highlighted :global(.hljs-name),
  .code-block.highlighted :global(.hljs-tag) {
    color: #569cd6;
  }

  .code-block.highlighted :global(.hljs-attribute) {
    color: #9cdcfe;
  }

  .code-block.highlighted :global(.hljs-symbol),
  .code-block.highlighted :global(.hljs-punctuation) {
    color: #d4d4d4;
  }

  /* Light theme adjustments */
  :global(.light) .code-block.highlighted :global(.hljs-string),
  :global(.light) .code-block.highlighted :global(.hljs-attr) {
    color: #a31515;
  }

  :global(.light) .code-block.highlighted :global(.hljs-number) {
    color: #098658;
  }

  :global(.light) .code-block.highlighted :global(.hljs-literal),
  :global(.light) .code-block.highlighted :global(.hljs-keyword) {
    color: #0000ff;
  }

  :global(.light) .code-block.highlighted :global(.hljs-name),
  :global(.light) .code-block.highlighted :global(.hljs-tag) {
    color: #800000;
  }

  :global(.light) .code-block.highlighted :global(.hljs-attribute) {
    color: #ff0000;
  }

  :global(.light) .code-block.highlighted :global(.hljs-symbol),
  :global(.light) .code-block.highlighted :global(.hljs-punctuation) {
    color: #1e1e1e;
  }
</style>

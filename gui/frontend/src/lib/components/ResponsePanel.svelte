<script lang="ts">
  import { onMount, onDestroy, afterUpdate } from 'svelte';
  import { currentResponse, responseTab, isExecuting } from '../stores/app';
  import {
    getStatusColor, formatHeaders, formatDuration, formatSize,
    getResponseSize, getHighlightedBody, applySearchHighlights,
  } from './ResponsePanel';
  import styles from './ResponsePanel.module.css';

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

<div class={styles['response-panel']}>
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
    <div class={styles.loading}>
      <div class={styles.spinner}></div>
      <p>Executing request...</p>
    </div>
  {:else if response}
    {#if response.error}
      <div class={styles.error}>
        <span class={styles['error-icon']}>⚠</span>
        <span class={styles['error-text']}>{response.error}</span>
      </div>
    {:else}
      <div class={styles['status-bar']}>
        <span class={styles.status} style="color: {getStatusColor(response.statusCode)}">
          {response.statusCode} {response.status.replace(String(response.statusCode), '').trim()}
        </span>
        <span class={styles.duration}>{formatDuration(response.duration)}</span>
        <span class={styles.size}>{formatSize(responseSize)}</span>
        <span class={styles['content-type']}>{response.contentType || 'Unknown'}</span>
        <button class={styles['copy-btn']} on:click={copyToClipboard} title="Copy response body">
          Copy
        </button>
      </div>

      {#if searchVisible && $responseTab === 'body'}
        <div class={styles['search-bar']}>
          <input
            bind:this={searchInput}
            bind:value={searchQuery}
            on:keydown={handleSearchKeydown}
            class={styles['search-input']}
            placeholder="Search…"
            autocomplete="off"
            spellcheck="false"
          />
          <span class={styles['match-count']}>
            {#if searchQuery.trim() && matchCount > 0}
              {matchIndex + 1} / {matchCount}
            {:else if searchQuery.trim()}
              No results
            {/if}
          </span>
          <button class={styles['search-nav-btn']} on:click={prevMatch} disabled={matchCount === 0} title="Previous (Shift+Enter)">↑</button>
          <button class={styles['search-nav-btn']} on:click={nextMatch} disabled={matchCount === 0} title="Next (Enter)">↓</button>
          <button class={styles['search-close-btn']} on:click={closeSearch} title="Close (Esc)">×</button>
        </div>
      {/if}

      <div class={styles.content}>
        {#if $responseTab === 'body'}
          <pre class="{styles['code-block']} {styles.highlighted}" bind:this={preElement}>{@html displayBody || 'Empty response'}</pre>
        {:else if $responseTab === 'headers'}
          <pre class={styles['code-block']} bind:this={headersPreElement}>{formatHeaders(response.headers) || 'No headers'}</pre>
        {/if}
      </div>
    {/if}
  {:else}
    <div class={styles.empty}>
      <p>No response yet</p>
      <p class={styles.hint}>Execute a request to see the response</p>
    </div>
  {/if}
</div>


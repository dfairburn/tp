<script lang="ts">
  import { currentResponse, responseTab, isExecuting } from '../stores/app';
  import hljs from 'highlight.js/lib/core';
  import json from 'highlight.js/lib/languages/json';
  import xml from 'highlight.js/lib/languages/xml';
  import graphql from 'highlight.js/lib/languages/graphql';
  
  // Register only the languages we need
  hljs.registerLanguage('json', json);
  hljs.registerLanguage('xml', xml);
  hljs.registerLanguage('graphql', graphql);

  $: response = $currentResponse;

  function getStatusColor(code: number): string {
    if (code >= 200 && code < 300) return '#49cc90';
    if (code >= 300 && code < 400) return '#fca130';
    if (code >= 400 && code < 500) return '#f93e3e';
    if (code >= 500) return '#f93e3e';
    return '#999';
  }

  function formatHeaders(headers: Record<string, string> | null | undefined): string {
    if (!headers) return '';
    return Object.entries(headers)
      .map(([k, v]) => `${k}: ${v}`)
      .join('\n');
  }

  function formatDuration(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
  }

  function getResponseSize(body: string | undefined): number {
    if (!body) return 0;
    return new Blob([body]).size;
  }

  async function copyToClipboard() {
    if (!response?.body) return;
    try {
      await navigator.clipboard.writeText(response.body);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  }

  function getHighlightedBody(body: string | undefined, contentType: string | undefined): string {
    if (!body) return '';
    
    const type = contentType?.toLowerCase() || '';
    
    try {
      if (type.includes('json')) {
        return hljs.highlight(body, { language: 'json' }).value;
      }
      
      if (type.includes('graphql')) {
        return hljs.highlight(body, { language: 'graphql' }).value;
      }
      
      if (type.includes('xml') || type.includes('html')) {
        return hljs.highlight(body, { language: 'xml' }).value;
      }
    } catch (e) {
      // Fall back to escaped plain text on error
      console.warn('Highlight error:', e);
    }
    
    // Plain text - just escape HTML
    return body
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
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

      <div class="content">
        {#if $responseTab === 'body'}
          <pre class="code-block highlighted">{@html highlightedBody || 'Empty response'}</pre>
        {:else if $responseTab === 'headers'}
          <pre class="code-block">{formatHeaders(response.headers) || 'No headers'}</pre>
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

<script lang="ts">
  import { selectedTemplate, requestTab, overrides, isExecuting, currentResponse, openTemplateEditor } from '../stores/app';
  import { ExecuteTemplateWithOverrides } from '../../../wailsjs/go/main/App';
  import hljs from 'highlight.js/lib/core';
  import json from 'highlight.js/lib/languages/json';
  import xml from 'highlight.js/lib/languages/xml';
  import graphql from 'highlight.js/lib/languages/graphql';
  
  // Register languages
  hljs.registerLanguage('json', json);
  hljs.registerLanguage('xml', xml);
  hljs.registerLanguage('graphql', graphql);

  $: template = $selectedTemplate;
  
  // Extract variables used in the template (from URL, headers, body)
  $: usedVariables = template ? extractVariables(template) : [];
  
  // Detect body language and highlight
  $: bodyLanguage = template?.body ? detectLanguage(template.body, template.headers) : null;
  $: highlightedBody = template?.body ? highlightBody(template.body, bodyLanguage) : '';
  
  function detectLanguage(body: string, headers: Record<string, string> | null | undefined): string | null {
    // Check Content-Type header first
    const contentType = headers?.['Content-Type'] || headers?.['content-type'] || '';
    if (contentType.includes('json')) return 'json';
    if (contentType.includes('xml')) return 'xml';
    if (contentType.includes('graphql')) return 'graphql';
    
    // Try to detect from body content
    const trimmed = body.trim();
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) return 'json';
    if (trimmed.startsWith('<')) return 'xml';
    if (trimmed.match(/^\s*(query|mutation|subscription|fragment)\s/)) return 'graphql';
    
    return null;
  }
  
  function highlightBody(body: string, language: string | null): string {
    if (!body) return '';
    
    if (language) {
      try {
        return hljs.highlight(body, { language }).value;
      } catch (e) {
        console.warn('Highlight error:', e);
      }
    }
    
    // Plain text - just escape HTML
    return body
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }
  
  function extractVariables(tmpl: typeof template): string[] {
    if (!tmpl) return [];
    const vars = new Set<string>();
    const regex = /\{\{\.(\w+)\}\}/g;
    
    // Check URL
    if (tmpl.url) {
      let match;
      while ((match = regex.exec(tmpl.url)) !== null) {
        vars.add(match[1]);
      }
    }
    
    // Check headers
    if (tmpl.headers) {
      for (const value of Object.values(tmpl.headers)) {
        let match;
        regex.lastIndex = 0;
        while ((match = regex.exec(String(value))) !== null) {
          vars.add(match[1]);
        }
      }
    }
    
    // Check body
    if (tmpl.body) {
      let match;
      regex.lastIndex = 0;
      while ((match = regex.exec(tmpl.body)) !== null) {
        vars.add(match[1]);
      }
    }
    
    return Array.from(vars).sort();
  }

  async function executeRequest() {
    if (!template) return;
    
    isExecuting.set(true);
    currentResponse.set(null);
    
    try {
      const response = await ExecuteTemplateWithOverrides(template.absolutePath, $overrides);
      currentResponse.set(response);
    } catch (err) {
      currentResponse.set({
        statusCode: 0,
        status: 'Error',
        headers: {},
        body: '',
        contentType: '',
        duration: 0,
        error: String(err),
      });
    } finally {
      isExecuting.set(false);
    }
  }

  async function openInEditor() {
    if (!template) return;
    openTemplateEditor(template);
  }

  function getMethodColor(method: string): string {
    const colors: Record<string, string> = {
      'GET': '#61affe',
      'POST': '#49cc90',
      'PUT': '#fca130',
      'PATCH': '#50e3c2',
      'DELETE': '#f93e3e',
      'HEAD': '#9012fe',
      'OPTIONS': '#0d5aa7',
    };
    return colors[method?.toUpperCase()] || '#999';
  }

  function formatHeaders(headers: Record<string, string> | null | undefined): string {
    if (!headers) return '';
    return Object.entries(headers)
      .map(([k, v]) => `${k}: ${v}`)
      .join('\n');
  }
</script>

<div class="request-panel">
  <div class="panel-header">
    <span class="title">Request</span>
    <div class="tabs">
      <button 
        class="tab" 
        class:active={$requestTab === 'headers'}
        on:click={() => requestTab.set('headers')}
      >
        Headers
      </button>
      <button 
        class="tab" 
        class:active={$requestTab === 'body'}
        on:click={() => requestTab.set('body')}
      >
        Body
      </button>
      <button 
        class="tab" 
        class:active={$requestTab === 'params'}
        on:click={() => requestTab.set('params')}
      >
        Params {#if usedVariables.length > 0}<span class="badge">{usedVariables.length}</span>{/if}
      </button>
    </div>
  </div>

  {#if template}
    <div class="url-bar">
      <span class="method" style="background: {getMethodColor(template.method)}">{template.method || 'GET'}</span>
      <span class="url">{template.url || 'No URL'}</span>
      <button class="execute-btn" on:click={executeRequest} disabled={$isExecuting}>
        {$isExecuting ? '⏳' : '▶'} Execute
      </button>
      <button class="edit-btn" on:click={openInEditor} title="Edit template">
        ✎
      </button>
    </div>

    <div class="content">
      {#if $requestTab === 'headers'}
        <pre class="code-block">{formatHeaders(template.headers) || 'No headers'}</pre>
      {:else if $requestTab === 'body'}
        <pre class="code-block highlighted">{@html highlightedBody || 'No body'}</pre>
      {:else if $requestTab === 'params'}
        <div class="params-list">
          {#if usedVariables.length === 0}
            <div class="params-empty">No variables used in this template</div>
          {:else}
            {#each usedVariables as varName}
              <div class="param-item">
                <div class="param-name">
                  <code>{'{{.'}{varName}{'}}'}</code>
                </div>
                <div class="param-desc">
                  {#if template.descriptions && template.descriptions[varName]}
                    {template.descriptions[varName]}
                  {:else}
                    <span class="no-desc">No description</span>
                  {/if}
                </div>
              </div>
            {/each}
          {/if}
        </div>
      {/if}
    </div>
  {:else}
    <div class="empty">
      <p>Select a template to view request details</p>
      <p class="hint">Use arrow keys or click to navigate</p>
    </div>
  {/if}
</div>

<style>
  .request-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-primary);
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

  .url-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .method {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    color: white;
    text-transform: uppercase;
  }

  .url {
    flex: 1;
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .execute-btn {
    padding: 6px 16px;
    border: none;
    border-radius: 4px;
    background: var(--accent-color);
    color: white;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s;
  }

  .execute-btn:hover:not(:disabled) {
    opacity: 0.9;
  }

  .execute-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .edit-btn {
    padding: 6px 10px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .edit-btn:hover {
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

  .badge {
    display: inline-block;
    padding: 1px 6px;
    margin-left: 4px;
    font-size: 10px;
    font-weight: 600;
    background: var(--accent-color);
    color: white;
    border-radius: 10px;
  }

  .params-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .params-empty {
    color: var(--text-muted);
    font-style: italic;
    padding: 12px 0;
  }

  .param-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
  }

  .param-name code {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 13px;
    color: var(--accent-color);
    background: var(--bg-hover);
    padding: 2px 6px;
    border-radius: 3px;
  }

  .param-desc {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .param-desc .no-desc {
    color: var(--text-muted);
    font-style: italic;
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

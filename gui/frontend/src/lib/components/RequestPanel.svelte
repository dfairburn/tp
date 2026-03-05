<script lang="ts">
  import { onMount } from 'svelte';
  import { selectedTemplate, requestTab, overrides, isExecuting, currentResponse, paramValuesCache } from '../stores/app';
  import { ExecuteTemplateWithOverrides, SaveTemplate, GetTemplate, PreviewTemplate, PreviewBody, GetVariables } from '../../../wailsjs/go/main/App';
  import yaml from 'js-yaml';
  import hljs from 'highlight.js/lib/core';
  import json from 'highlight.js/lib/languages/json';
  import xml from 'highlight.js/lib/languages/xml';
  import graphql from 'highlight.js/lib/languages/graphql';

  hljs.registerLanguage('json', json);
  hljs.registerLanguage('xml', xml);
  hljs.registerLanguage('graphql', graphql);

  const HTTP_METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];

  let envVars: Record<string, any> = {};
  onMount(async () => {
    try { envVars = await GetVariables() || {}; } catch {}
  });

  $: template = $selectedTemplate;

  // Exit edit mode and clear preview when template selection changes
  let prevPath = '';
  $: {
    const path = $selectedTemplate?.absolutePath ?? '';
    if (path !== prevPath) {
      prevPath = path;
      editing = false;
      saveError = '';
      rawMode = false;
      rawBody = '';
      saveRawError = '';
      previewUrl = '';
      previewBody = '';
      previewHeaders = {};
      previewError = '';
    }
  }

  // View mode derived state
  $: usedVariables = template ? extractVariables(template) : [];
  $: headerVariables = template ? extractHeaderVars(template) : [];
  $: bodyAndUrlVariables = template ? extractBodyAndUrlVars(template) : [];
  $: bodyLanguage = template?.body ? detectLanguage(template.body, template.headers) : null;
  $: highlightedBody = template?.body ? highlightBody(template.body, bodyLanguage) : '';
  // Go-rendered preview state
  let previewUrl = '';
  let previewBody = '';
  let previewHeaders: Record<string, string> = {};
  let previewError = '';
  let previewTimer: ReturnType<typeof setTimeout> | null = null;
  $: highlightedPreviewBody = previewBody ? highlightBody(formatBody(previewBody, bodyLanguage), bodyLanguage) : '';
  $: displayHeaders = Object.keys(previewHeaders).length > 0 ? previewHeaders : (template?.headers || {});

  // Raw body editing (view mode inline editor)
  let rawMode = false;
  let rawBody = '';
  let saveRawError = '';

  function toggleRaw() {
    rawMode = !rawMode;
    if (rawMode) {
      rawBody = template?.body || '';
      saveRawError = '';
    }
  }

  async function saveBodyOnly() {
    if (!template) return;
    saveRawError = '';
    try {
      const headers: Record<string, string> = {};
      for (const [k, v] of Object.entries(template.headers || {})) {
        headers[k] = String(v);
      }
      const descriptions: Record<string, string> = {};
      for (const [k, v] of Object.entries(template.descriptions || {})) {
        descriptions[k] = String(v);
      }
      const templateObj: Record<string, any> = {
        name: template.name,
        method: template.method,
        url: template.url,
        headers,
        body: rawBody,
      };
      if (Object.keys(descriptions).length > 0) templateObj.descriptions = descriptions;
      await SaveTemplate(template.absolutePath, yaml.dump(templateObj));
      const updated = await GetTemplate(template.absolutePath);
      selectedTemplate.set(updated);
    } catch (err) {
      saveRawError = String(err);
    }
  }

  // Edit mode state
  let editing = false;
  let editMethod = 'GET';
  let editUrl = '';
  let editHeaders: { key: string; value: string }[] = [{ key: '', value: '' }];
  let editBody = '';
  let editDescription = '';
  let editParamDescriptions: Record<string, string> = {};
  let saveError = '';

  $: editVariables = editing ? extractVarsFromStrings(editUrl, editHeaders, editBody) : [];

  function enterEditMode() {
    if (!template) return;
    editMethod = template.method || 'GET';
    editUrl = template.url || '';
    const ha = Object.entries(template.headers || {}).map(([k, v]) => ({ key: k, value: String(v) }));
    editHeaders = ha.length > 0 ? ha : [{ key: '', value: '' }];
    editBody = template.body || '';
    editDescription = template.descriptions?.description || '';
    editParamDescriptions = {};
    for (const v of usedVariables) {
      editParamDescriptions[v] = template.descriptions?.[v] || '';
    }
    saveError = '';
    editing = true;
  }

  function cancelEdit() {
    editing = false;
    saveError = '';
  }

  async function saveEdit() {
    if (!template) return;
    saveError = '';
    try {
      const headers: Record<string, string> = {};
      for (const h of editHeaders) {
        if (h.key.trim()) headers[h.key.trim()] = h.value;
      }

      const descriptions: Record<string, string> = {};
      if (editDescription.trim()) descriptions['description'] = editDescription.trim();
      for (const v of extractVarsFromStrings(editUrl, editHeaders, editBody)) {
        if (editParamDescriptions[v]?.trim()) {
          descriptions[v] = editParamDescriptions[v].trim();
        }
      }

      const templateObj: Record<string, any> = {
        name: template.name,
        method: editMethod,
        url: editUrl,
        headers,
        body: editBody,
      };
      if (Object.keys(descriptions).length > 0) {
        templateObj.descriptions = descriptions;
      }

      const yamlContent = yaml.dump(templateObj);
      await SaveTemplate(template.absolutePath, yamlContent);
      const updated = await GetTemplate(template.absolutePath);
      selectedTemplate.set(updated);
      editing = false;
    } catch (err) {
      saveError = String(err);
    }
  }

  function addEditHeader() {
    editHeaders = [...editHeaders, { key: '', value: '' }];
  }

  function removeEditHeader(i: number) {
    editHeaders = editHeaders.filter((_, idx) => idx !== i);
    if (editHeaders.length === 0) editHeaders = [{ key: '', value: '' }];
  }

  function updateParamValue(varName: string, value: string) {
    overrides.update(o => ({ ...o, [varName]: value }));
    if (template) {
      paramValuesCache.update(cache => ({
        ...cache,
        [template.absolutePath]: { ...(cache[template.absolutePath] || {}), [varName]: value },
      }));
    }
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

  function getMethodColor(method: string): string {
    const colors: Record<string, string> = {
      GET: '#61affe',
      POST: '#49cc90',
      PUT: '#fca130',
      PATCH: '#50e3c2',
      DELETE: '#f93e3e',
      HEAD: '#9012fe',
      OPTIONS: '#0d5aa7',
    };
    return colors[method?.toUpperCase()] || '#999';
  }

  function formatHeaders(headers: Record<string, string> | null | undefined): string {
    if (!headers) return '';
    return Object.entries(headers)
      .map(([k, v]) => `${k}: ${v}`)
      .join('\n');
  }

  function detectLanguage(body: string, headers: Record<string, string> | null | undefined): string | null {
    const contentType = headers?.['Content-Type'] || headers?.['content-type'] || '';
    if (contentType.includes('json')) return 'json';
    if (contentType.includes('xml')) return 'xml';
    if (contentType.includes('graphql')) return 'graphql';
    const trimmed = body.trim();
    if (trimmed.startsWith('{') || trimmed.startsWith('[')) return 'json';
    if (trimmed.startsWith('<')) return 'xml';
    if (trimmed.match(/^\s*(query|mutation|subscription|fragment)\s/)) return 'graphql';
    return null;
  }

  function formatBody(body: string, language: string | null): string {
    if (language === 'json' && body) {
      try {
        return JSON.stringify(JSON.parse(body), null, 2);
      } catch {
        // Not valid JSON (may contain template vars or be malformed) — leave as-is
      }
    }
    return body;
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
    return body
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
  }

  async function doPreview() {
    if (!template || editing) return;
    if (rawMode) {
      const result = await PreviewBody(rawBody, $overrides);
      previewBody = result.body || '';
      previewError = result.error || '';
    } else {
      const result = await PreviewTemplate(template.absolutePath, $overrides);
      previewUrl = result.url || '';
      previewBody = result.body || '';
      previewHeaders = result.headers || {};
      previewError = result.error || '';
    }
  }

  function schedulePreview() {
    if (previewTimer) clearTimeout(previewTimer);
    previewTimer = setTimeout(doPreview, 200);
  }

  // Trigger preview whenever template or overrides change (not while editing).
  // Also track rawBody when in raw mode for live preview.
  $: if (!editing) { template; $overrides; if (rawMode) rawBody; schedulePreview(); }

  // Extract variable name from a single {{...}} block inner content,
  // matching the same patterns as handlers/use-help.go ParseUsages.
  function extractVarFromBlock(inner: string): string | null {
    let m: RegExpMatchArray | null;
    if ((m = inner.match(/^\s*\.(\S+)\s*$/)))                           return m[1]; // {{.name}}
    if ((m = inner.match(/^\s*optional.*\.(\S+)\s*$/)))                 return m[1]; // {{optional ... .name}}
    if ((m = inner.match(/^\s*timestamp\s+\.(\S+)\s*$/)))               return m[1]; // {{timestamp .name}}
    if ((m = inner.match(/^\s*default\s+\.(\S+)\s+"[^"]*"\s*$/)))       return m[1]; // {{default .name "val"}}
    if ((m = inner.match(/^\s*default\s+"[^"]*"\s+\.(\S+)\s*$/)))       return m[1]; // {{default "val" .name}}
    return null;
  }

  function extractVarsFromStr(str: string, vars: Set<string>) {
    const blockRegex = /\{\{([^}]+)\}\}/g;
    let block: RegExpExecArray | null;
    while ((block = blockRegex.exec(str)) !== null) {
      const v = extractVarFromBlock(block[1]);
      if (v) vars.add(v);
    }
  }

  function extractVariables(tmpl: typeof template): string[] {
    if (!tmpl) return [];
    const vars = new Set<string>();
    if (tmpl.url) extractVarsFromStr(tmpl.url, vars);
    if (tmpl.headers) {
      for (const value of Object.values(tmpl.headers)) extractVarsFromStr(String(value), vars);
    }
    if (tmpl.body) extractVarsFromStr(tmpl.body, vars);
    return Array.from(vars).sort();
  }

  function extractHeaderVars(tmpl: typeof template): string[] {
    if (!tmpl?.headers) return [];
    const vars = new Set<string>();
    for (const value of Object.values(tmpl.headers)) extractVarsFromStr(String(value), vars);
    return Array.from(vars).sort();
  }

  function extractBodyAndUrlVars(tmpl: typeof template): string[] {
    if (!tmpl) return [];
    const vars = new Set<string>();
    if (tmpl.url) extractVarsFromStr(tmpl.url, vars);
    if (tmpl.body) extractVarsFromStr(tmpl.body, vars);
    return Array.from(vars).sort();
  }

  function extractVarsFromStrings(url: string, headers: { key: string; value: string }[], body: string): string[] {
    const vars = new Set<string>();
    extractVarsFromStr(url, vars);
    for (const h of headers) extractVarsFromStr(h.value, vars);
    extractVarsFromStr(body, vars);
    return Array.from(vars).sort();
  }
</script>

<div class="request-panel">
  <div class="panel-header">
    <span class="title">{editing ? 'Edit Template' : 'Request'}</span>
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
    </div>
  </div>

  {#if template}
    <div class="url-bar">
      {#if editing}
        <select bind:value={editMethod} class="method-select">
          {#each HTTP_METHODS as m}
            <option value={m}>{m}</option>
          {/each}
        </select>
        <input
          type="text"
          bind:value={editUrl}
          class="url-input"
          placeholder="https://..."
          autocomplete="off"
          spellcheck="false"
        />
        <button class="cancel-btn" on:click={cancelEdit}>Cancel</button>
        <button class="save-btn" on:click={saveEdit}>Save</button>
      {:else}
        <span class="method" style="background: {getMethodColor(template.method)}">{template.method || 'GET'}</span>
        <span class="url" title={template.url}>{previewUrl || template.url || 'No URL'}</span>
        <button class="execute-btn" on:click={executeRequest} disabled={$isExecuting}>
          {$isExecuting ? '⏳' : '▶'} Execute
        </button>
        <button class="edit-btn" on:click={enterEditMode} title="Edit template">✎</button>
      {/if}
    </div>

    {#if editing}
      <div class="description-bar">
        <span class="desc-label">Description</span>
        <input
          type="text"
          bind:value={editDescription}
          class="description-input"
          placeholder="Template description (optional)"
          autocomplete="off"
        />
      </div>
    {/if}

    {#if saveError}
      <div class="error-bar">{saveError}</div>
    {/if}

    <div class="content">
      {#if $requestTab === 'headers'}
        {#if editing}
          <div class="edit-section">
            {#each editHeaders as _, i}
              <div class="header-row">
                <input
                  type="text"
                  value={editHeaders[i].key}
                  on:input={e => { editHeaders[i].key = e.currentTarget.value; editHeaders = editHeaders; }}
                  placeholder="Header name"
                  autocomplete="off"
                  spellcheck="false"
                />
                <input
                  type="text"
                  value={editHeaders[i].value}
                  on:input={e => { editHeaders[i].value = e.currentTarget.value; editHeaders = editHeaders; }}
                  placeholder="Value"
                  autocomplete="off"
                  spellcheck="false"
                />
                <button class="remove-btn" on:click={() => removeEditHeader(i)}>×</button>
              </div>
            {/each}
            <button class="add-btn" on:click={addEditHeader}>+ Add Header</button>
          </div>
        {:else}
          {#if headerVariables.length > 0}
            <div class="params-grid">
              {#each headerVariables as varName}
                <code class="pgrid-key">{'{{.'}{varName}{'}}'}</code>
                <input
                  type="text"
                  class="pgrid-value"
                  value={$overrides[varName] ?? ''}
                  on:input={e => updateParamValue(varName, e.currentTarget.value)}
                  placeholder={envVars[varName] != null ? String(envVars[varName]) : 'Value...'}
                  autocomplete="off"
                  spellcheck="false"
                />
                <span class="pgrid-desc">{template.descriptions?.[varName] ?? ''}</span>
              {/each}
            </div>
          {/if}
          {#if Object.keys(displayHeaders).length > 0}
            <div class="headers-grid">
              {#each Object.entries(displayHeaders) as [key, value]}
                <span class="hgrid-key">{key}</span>
                <span class="hgrid-value">{value}</span>
              {/each}
            </div>
          {:else}
            <p class="no-content">No headers</p>
          {/if}
        {/if}

      {:else if $requestTab === 'body'}
        {#if editing}
          <textarea
            class="body-edit"
            bind:value={editBody}
            spellcheck="false"
            placeholder="Request body..."
          ></textarea>
          {#if editVariables.length > 0}
            <div class="edit-params-section">
              <span class="section-label">Param descriptions</span>
              {#each editVariables as varName}
                <div class="edit-param-row">
                  <code>{'{{.'}{varName}{'}}'}</code>
                  <input
                    type="text"
                    class="param-desc-input"
                    value={editParamDescriptions[varName] ?? ''}
                    on:input={e => editParamDescriptions[varName] = e.currentTarget.value}
                    placeholder="Description (optional)"
                    autocomplete="off"
                  />
                </div>
              {/each}
            </div>
          {/if}
        {:else}
          {#if bodyAndUrlVariables.length > 0}
            <div class="params-grid">
              {#each bodyAndUrlVariables as varName}
                <code class="pgrid-key">{'{{.'}{varName}{'}}'}</code>
                <input
                  type="text"
                  class="pgrid-value"
                  value={$overrides[varName] ?? ''}
                  on:input={e => updateParamValue(varName, e.currentTarget.value)}
                  placeholder={envVars[varName] != null ? String(envVars[varName]) : 'Value...'}
                  autocomplete="off"
                  spellcheck="false"
                />
                <span class="pgrid-desc">{template.descriptions?.[varName] ?? ''}</span>
              {/each}
            </div>
          {/if}

          <div class="body-toolbar">
            <button class="raw-toggle" class:active={rawMode} on:click={toggleRaw} title="Edit raw template">
              ✎ Raw
            </button>
          </div>

          {#if previewError}
            <div class="preview-error">{previewError}</div>
          {/if}

          {#if rawMode}
            <div class="body-split">
              <div class="body-split-pane">
                <pre class="code-block highlighted split-pre">{@html (previewBody ? highlightedPreviewBody : highlightedBody) || 'No body'}</pre>
              </div>
              <div class="body-split-pane body-split-right">
                <textarea
                  class="body-raw-edit"
                  bind:value={rawBody}
                  spellcheck="false"
                  placeholder="Request body..."
                ></textarea>
                <div class="raw-actions">
                  {#if saveRawError}<span class="preview-error">{saveRawError}</span>{/if}
                  <button class="save-raw-btn" on:click={saveBodyOnly}>Save</button>
                </div>
              </div>
            </div>
          {:else}
            <pre class="code-block highlighted">{@html (previewBody ? highlightedPreviewBody : highlightedBody) || 'No body'}</pre>
          {/if}
        {/if}
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

  /* View mode url-bar */
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
    white-space: nowrap;
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
    white-space: nowrap;
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

  /* Edit mode url-bar elements */
  .method-select {
    padding: 5px 6px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    outline: none;
  }

  .method-select:focus {
    border-color: var(--accent-color);
  }

  .url-input {
    flex: 1;
    padding: 5px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    outline: none;
    min-width: 0;
  }

  .url-input:focus {
    border-color: var(--accent-color);
  }

  .cancel-btn {
    padding: 6px 14px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
    white-space: nowrap;
  }

  .cancel-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .save-btn {
    padding: 6px 16px;
    border: none;
    border-radius: 4px;
    background: var(--accent-color);
    color: white;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s;
    white-space: nowrap;
  }

  .save-btn:hover {
    opacity: 0.9;
  }

  /* Description bar (edit mode only) */
  .description-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-bottom: 1px solid var(--border-color);
    background: var(--bg-secondary);
  }

  .desc-label {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
    white-space: nowrap;
  }

  .description-input {
    flex: 1;
    padding: 4px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 12px;
    outline: none;
  }

  .description-input:focus {
    border-color: var(--accent-color);
  }

  /* Error bar (save errors) */
  .error-bar {
    padding: 6px 12px;
    background: rgba(249, 62, 62, 0.1);
    border-bottom: 1px solid rgba(249, 62, 62, 0.3);
    color: #f93e3e;
    font-size: 12px;
  }

  /* Preview render error (inline, inside content area) */
  .preview-error {
    padding: 4px 8px;
    margin-bottom: 8px;
    background: rgba(249, 62, 62, 0.08);
    border: 1px solid rgba(249, 62, 62, 0.25);
    border-radius: 4px;
    color: #f93e3e;
    font-size: 11px;
  }

  /* Content area */
  .content {
    flex: 1;
    overflow: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
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

  /* Edit mode: headers */
  .edit-section {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .header-row {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .header-row input {
    flex: 1;
    padding: 6px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    outline: none;
    min-width: 0;
  }

  .header-row input:focus {
    border-color: var(--accent-color);
  }

  .remove-btn {
    padding: 4px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-secondary);
    color: var(--text-muted);
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s;
    line-height: 1;
  }

  .remove-btn:hover {
    background: #f93e3e;
    border-color: #f93e3e;
    color: white;
  }

  .add-btn {
    align-self: flex-start;
    margin-top: 4px;
    padding: 5px 12px;
    border: 1px dashed var(--border-color);
    border-radius: 4px;
    background: transparent;
    color: var(--text-secondary);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .add-btn:hover {
    border-color: var(--accent-color);
    color: var(--accent-color);
  }

  /* Edit mode: body */
  .body-edit {
    flex: 1;
    width: 100%;
    min-height: 200px;
    padding: 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    line-height: 1.5;
    resize: vertical;
    outline: none;
    box-sizing: border-box;
  }

  .body-edit:focus {
    border-color: var(--accent-color);
  }

  /* View mode: params grid (key | value | description) */
  .params-grid {
    display: grid;
    grid-template-columns: auto 1fr minmax(0, 200px);
    align-items: center;
    gap: 4px 8px;
    padding-bottom: 10px;
    margin-bottom: 10px;
    border-bottom: 1px solid var(--border-color);
  }

  .pgrid-key {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    color: var(--accent-color);
    background: var(--bg-hover);
    padding: 2px 6px;
    border-radius: 3px;
    white-space: nowrap;
  }

  .pgrid-value {
    padding: 4px 7px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    outline: none;
    min-width: 0;
  }

  .pgrid-value:focus {
    border-color: var(--accent-color);
  }

  .pgrid-desc {
    font-size: 11px;
    color: var(--text-muted);
    font-style: italic;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Raw body editing */
  .body-toolbar {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 6px;
  }

  .raw-toggle {
    padding: 3px 10px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-secondary);
    color: var(--text-secondary);
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .raw-toggle:hover {
    border-color: var(--accent-color);
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  .raw-toggle.active {
    background: var(--accent-color);
    border-color: var(--accent-color);
    color: white;
  }

  /* Side-by-side split */
  .body-split {
    display: flex;
    gap: 8px;
    flex: 1;
    min-height: 200px;
  }

  .body-split-pane {
    flex: 1;
    min-width: 0;
    overflow: auto;
  }

  .body-split-right {
    display: flex;
    flex-direction: column;
    gap: 6px;
    overflow: visible;
  }

  .split-pre {
    height: 100%;
    overflow: auto;
  }

  .body-raw-edit {
    flex: 1;
    width: 100%;
    min-height: 120px;
    padding: 8px;
    border: 1px solid var(--accent-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    line-height: 1.5;
    resize: none;
    outline: none;
    box-sizing: border-box;
  }

  .raw-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 8px;
    flex-shrink: 0;
  }

  .save-raw-btn {
    padding: 5px 16px;
    border: none;
    border-radius: 4px;
    background: var(--accent-color);
    color: white;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s;
  }

  .save-raw-btn:hover {
    opacity: 0.9;
  }

  /* View mode: rendered headers grid */
  .headers-grid {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 4px 12px;
    align-items: baseline;
  }

  .hgrid-key {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    color: var(--text-muted);
    white-space: nowrap;
  }

  .hgrid-value {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    color: var(--text-primary);
    word-break: break-all;
  }

  .no-content {
    margin: 0;
    font-size: 12px;
    color: var(--text-muted);
  }

  /* Edit mode: param descriptions beneath body textarea */
  .edit-params-section {
    display: flex;
    flex-direction: column;
    gap: 5px;
    padding-top: 8px;
    margin-top: 6px;
    border-top: 1px solid var(--border-color);
  }

  .section-label {
    font-size: 10px;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    margin-bottom: 2px;
  }

  .edit-param-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .edit-param-row code {
    font-family: 'SF Mono', Monaco, 'Cascadia Code', monospace;
    font-size: 12px;
    color: var(--accent-color);
    background: var(--bg-hover);
    padding: 2px 6px;
    border-radius: 3px;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .param-desc-input {
    flex: 1;
    padding: 4px 7px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 12px;
    outline: none;
    min-width: 0;
  }

  .param-desc-input:focus {
    border-color: var(--accent-color);
  }

  /* Empty state */
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

  /* highlight.js syntax highlighting */
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

<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { get } from 'svelte/store';
  import { selectedTemplate, requestTab, overrides, isExecuting, currentResponse, paramValuesCache, bodyCache } from '../stores/app';
  import { ExecuteTemplateWithBodyAndOverrides, SaveTemplate, GetTemplate, PreviewTemplate, PreviewBody, GetVariables, RefreshVariables } from '../../../wailsjs/go/app/App';
  import yaml from 'js-yaml';
  import { getMethodColor, makeErrorResponse } from '../utils';
  import {
    HTTP_METHODS, BODYLESS_METHODS,
    extractVariables, extractHeaderVars, extractUrlOnlyVars, extractVarsFromStrings, extractBodyVarsOnly,
  } from './RequestPanel';
  import styles from './RequestPanel.module.css';
  import ParamRow from './ParamRow.svelte';

  let envVars: Record<string, any> = {};
  let isRefreshing = false;

  onMount(async () => {
    try { envVars = await GetVariables() || {}; } catch {}
  });

  async function refreshEnvVars() {
    isRefreshing = true;
    try {
      envVars = await RefreshVariables() || {};
      // Re-run preview so rendered headers/URL reflect the new values
      await doPreview();
    } catch {}
    isRefreshing = false;
  }

  $: template = $selectedTemplate;

  // Exit edit mode and clear preview when template selection changes
  let prevPath = '';
  $: {
    const path = $selectedTemplate?.absolutePath ?? '';
    if (path !== prevPath) {
      prevPath = path;
      editing = false;
      saveError = '';
      const cached = get(bodyCache);
      rawBody = (path && cached[path] !== undefined) ? cached[path] : ($selectedTemplate?.body || '');
      rawBodyTemplate = $selectedTemplate?.body || '';
      previewUrl = '';
      previewHeaders = {};
      previewError = '';
    }
  }

  // View mode derived state
  $: usedVariables = template ? extractVariables(template) : [];
  $: headerVariables = template ? extractHeaderVars(template) : [];
  $: urlVars = template ? extractUrlOnlyVars(template) : [];
  $: methodHasBody = !BODYLESS_METHODS.has((template?.method || 'GET').toUpperCase());
  $: bodyVars = template ? extractBodyVarsOnly(template) : [];
  $: hasBodyTemplate = bodyVars.length > 0;
  // Go-rendered preview state (URL only — body is always directly editable)
  let previewUrl = '';
  let previewHeaders: Record<string, string> = {};
  let previewError = '';
  let previewTimer: ReturnType<typeof setTimeout> | null = null;
  $: displayHeaders = Object.keys(previewHeaders).length > 0 ? previewHeaders : (template?.headers || {});

  // Raw body editing — persisted per-template in session (localStorage), not saved to file
  let rawBody = '';
  // Template raw body (the {{.varName}} template text, separate from the rendered rawBody)
  let rawBodyTemplate = '';

  // Template panel split/resize state
  let splitPercent = 50;
  let isDragging = false;
  let splitContainer: HTMLElement | null = null;

  function startDrag(e: MouseEvent) {
    isDragging = true;
    e.preventDefault();
    window.addEventListener('mousemove', onDragMove);
    window.addEventListener('mouseup', stopDrag);
  }

  function onDragMove(e: MouseEvent) {
    if (!splitContainer) return;
    const rect = splitContainer.getBoundingClientRect();
    const x = e.clientX - rect.left;
    splitPercent = Math.max(20, Math.min(80, (x / rect.width) * 100));
  }

  function stopDrag() {
    isDragging = false;
    window.removeEventListener('mousemove', onDragMove);
    window.removeEventListener('mouseup', stopDrag);
  }

  onDestroy(() => {
    window.removeEventListener('mousemove', onDragMove);
    window.removeEventListener('mouseup', stopDrag);
  });

  // Debounce bodyCache writes: avoid a localStorage write on every keystroke
  let _bodyCacheTimer: ReturnType<typeof setTimeout> | null = null;
  $: if (template) {
    const path = template.absolutePath;
    const body = rawBody;
    if (_bodyCacheTimer) clearTimeout(_bodyCacheTimer);
    _bodyCacheTimer = setTimeout(() => {
      bodyCache.update(c => ({ ...c, [path]: body }));
    }, 400);
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
      const response = await ExecuteTemplateWithBodyAndOverrides(template.absolutePath, $overrides, rawBody);
      currentResponse.set(response);
    } catch (err) {
      currentResponse.set(makeErrorResponse(err));
    } finally {
      isExecuting.set(false);
    }
  }

  async function doPreview() {
    if (!template || editing) return;
    const result = await PreviewTemplate(template.absolutePath, $overrides);
    previewUrl = result.url || '';
    previewHeaders = result.headers || {};
    previewError = result.error || '';
  }

  function schedulePreview() {
    if (previewTimer) clearTimeout(previewTimer);
    previewTimer = setTimeout(doPreview, 200);
  }

  // Trigger preview whenever template or overrides change (not while editing).
  $: if (!editing) { template; $overrides; schedulePreview(); }

  async function generateBody() {
    if (!rawBodyTemplate) return;
    previewError = '';
    try {
      const result = await PreviewBody(rawBodyTemplate, $overrides);
      if (result.error) {
        previewError = result.error;
      } else {
        rawBody = result.body;
      }
    } catch (err) {
      previewError = String(err);
    }
  }

  async function saveBodyTemplate() {
    if (!template) return;
    previewError = '';
    try {
      const headers: Record<string, string> = {};
      for (const [k, v] of Object.entries(template.headers || {})) {
        headers[k] = String(v);
      }
      const templateObj: Record<string, any> = {
        method: template.method,
        url: template.url,
        headers,
        body: rawBodyTemplate,
      };
      const descriptions = template.descriptions || {};
      if (Object.keys(descriptions).length > 0) {
        templateObj.descriptions = descriptions;
      }
      const yamlContent = yaml.dump(templateObj);
      await SaveTemplate(template.absolutePath, yamlContent);
      const updated = await GetTemplate(template.absolutePath);
      selectedTemplate.set(updated);
      rawBodyTemplate = updated.body || '';
    } catch (err) {
      previewError = String(err);
    }
  }


</script>

<div class={styles['request-panel']}>
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
    <div class={styles['url-bar']}>
      {#if editing}
        <select bind:value={editMethod} class={styles['method-select']}>
          {#each HTTP_METHODS as m}
            <option value={m}>{m}</option>
          {/each}
        </select>
        <input
          type="text"
          bind:value={editUrl}
          class={styles['url-input']}
          placeholder="https://..."
          autocomplete="off"
          spellcheck="false"
        />
        <button class={styles['cancel-btn']} on:click={cancelEdit}>Cancel</button>
        <button class={styles['save-btn']} on:click={saveEdit}>Save</button>
      {:else}
        <span class={styles.method} style="background: {getMethodColor(template.method)}">{template.method || 'GET'}</span>
        <span class={styles.url} title={template.url}>{previewUrl || template.url || 'No URL'}</span>
        <button class={styles['execute-btn']} on:click={executeRequest} disabled={$isExecuting}>
          {$isExecuting ? '⏳' : '▶'} Execute
        </button>
        <button class={styles['edit-btn']} on:click={enterEditMode} title="Edit template">✎</button>
        <button class={styles['refresh-btn']} on:click={refreshEnvVars} disabled={isRefreshing} title="Refresh environment variables">
          {isRefreshing ? '⏳' : '↻'}
        </button>
      {/if}
    </div>

    {#if editing}
      <div class={styles['description-bar']}>
        <span class={styles['desc-label']}>Description</span>
        <input
          type="text"
          bind:value={editDescription}
          class={styles['description-input']}
          placeholder="Template description (optional)"
          autocomplete="off"
        />
      </div>
    {/if}

    {#if saveError}
      <div class={styles['error-bar']}>{saveError}</div>
    {/if}

    <div class={styles.content}>
      {#if $requestTab === 'headers'}
        {#if editing}
          <div class={styles['edit-section']}>
            {#each editHeaders as _, i}
              <div class={styles['header-row']}>
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
                <button class={styles['remove-btn']} on:click={() => removeEditHeader(i)}>×</button>
              </div>
            {/each}
            <button class={styles['add-btn']} on:click={addEditHeader}>+ Add Header</button>
          </div>
        {:else}
          {#if headerVariables.length > 0}
            <div class={styles['params-grid']}>
              {#each headerVariables as varName}
                <ParamRow
                  {varName}
                  value={$overrides[varName] ?? ''}
                  placeholder={envVars[varName] != null ? String(envVars[varName]) : 'Value...'}
                  description={template.descriptions?.[varName] ?? ''}
                  on:change={e => updateParamValue(varName, e.detail)}
                />
              {/each}
            </div>
          {/if}
          {#if Object.keys(displayHeaders).length > 0}
            <div class={styles['headers-grid']}>
              {#each Object.entries(displayHeaders) as [key, value]}
                <span class={styles['hgrid-key']}>{key}</span>
                <span class={styles['hgrid-value']}>{value}</span>
              {/each}
            </div>
          {:else}
            <p class={styles['no-content']}>No headers</p>
          {/if}
        {/if}

      {:else if $requestTab === 'body'}
        {#if editing}
          <textarea
            class={styles['body-edit']}
            bind:value={editBody}
            spellcheck="false"
            placeholder="Request body..."
          ></textarea>
          {#if editVariables.length > 0}
            <div class={styles['edit-params-section']}>
              <span class={styles['section-label']}>Param descriptions</span>
              {#each editVariables as varName}
                <div class={styles['edit-param-row']}>
                  <code>{'{{.'}{varName}{'}}'}</code>
                  <input
                    type="text"
                    class={styles['param-desc-input']}
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
          {#if urlVars.length > 0}
            <div class={styles['params-grid']}>
              {#each urlVars as varName}
                <ParamRow
                  {varName}
                  value={$overrides[varName] ?? ''}
                  placeholder={envVars[varName] != null ? String(envVars[varName]) : 'Value...'}
                  description={template.descriptions?.[varName] ?? ''}
                  on:change={e => updateParamValue(varName, e.detail)}
                />
              {/each}
            </div>
          {/if}

          {#if methodHasBody}
            {#if previewError}
              <div class={styles['preview-error']}>{previewError}</div>
            {/if}

            <div class={styles['body-template-split']} bind:this={splitContainer}>
              <textarea
                class="{styles['body-raw-edit']} {styles['body-direct-edit']} {styles['split-body']}"
                style="flex: 0 0 {splitPercent}%"
                bind:value={rawBody}
                spellcheck="false"
                placeholder="Request body..."
              ></textarea>
              <div
                class="{styles['resize-handle']} {isDragging ? styles.dragging : ''}"
                on:mousedown={startDrag}
                role="separator"
                aria-label="Resize panels"
              ></div>
              <div class={styles['template-side-panel']}>
                <div class={styles['template-panel-header']}>
                  <span class={styles['section-label']}>Template</span>
                  <button class={styles['save-template-btn']} on:click={saveBodyTemplate} title="Save template body to file">Save</button>
                </div>
                {#if bodyVars.length > 0}
                  <div class={styles['template-vars-grid']}>
                    {#each bodyVars as varName}
                      <ParamRow
                        {varName}
                        value={$overrides[varName] ?? ''}
                        placeholder={envVars[varName] != null ? String(envVars[varName]) : 'Value...'}
                        on:change={e => updateParamValue(varName, e.detail)}
                      />
                    {/each}
                  </div>
                {:else}
                  <p class={styles['template-hint']}>Add {'{{.varName}}'} syntax to the template body to create input variables.</p>
                {/if}
                <textarea
                  class={styles['template-raw-edit']}
                  bind:value={rawBodyTemplate}
                  spellcheck="false"
                  placeholder="Template body..."
                ></textarea>
                <button class={styles['generate-btn']} on:click={generateBody}>← Generate body</button>
              </div>
            </div>
          {/if}
        {/if}
      {/if}
    </div>
  {:else}
    <div class={styles.empty}>
      <p>Select a template to view request details</p>
      <p class={styles.hint}>Use arrow keys or click to navigate</p>
    </div>
  {/if}
</div>


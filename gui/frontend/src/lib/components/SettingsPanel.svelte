<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { 
    GetVariables, 
    RefreshVariables, 
    GetEnvironmentFilePath,
    OpenEnvironmentFile,
    GetOverrides,
    OpenOverridesFile,
    SaveOverrides
  } from '../../../wailsjs/go/main/App';

  export let isOpen = false;

  const dispatch = createEventDispatcher();

  let variables: Record<string, any> = {};
  let overrides: Record<string, string> = {};
  let mergedVariables: Record<string, { value: any; source: 'env' | 'override' }> = {};
  let envFilePath = '';
  let isRefreshing = false;
  let isSaving = false;
  let statusMessage = '';
  
  // Inline editing state
  let editingKey: string | null = null;
  let editValue = '';
  let newKey = '';
  let newValue = '';

  $: if (isOpen) {
    loadData();
  }

  async function loadData() {
    try {
      const [vars, ovr, path] = await Promise.all([
        GetVariables(),
        GetOverrides(),
        GetEnvironmentFilePath()
      ]);
      variables = vars || {};
      overrides = ovr || {};
      envFilePath = path;
      mergeVariables();
    } catch (err) {
      console.error('Failed to load variables:', err);
      statusMessage = `Error loading: ${err}`;
    }
  }

  function mergeVariables() {
    mergedVariables = {};
    
    // Add env variables first
    for (const [key, value] of Object.entries(variables)) {
      mergedVariables[key] = { value, source: 'env' };
    }
    
    // Overrides take precedence
    for (const [key, value] of Object.entries(overrides)) {
      mergedVariables[key] = { value, source: 'override' };
    }
  }

  async function refreshVariables() {
    isRefreshing = true;
    statusMessage = '';
    try {
      const [vars, ovr] = await Promise.all([
        RefreshVariables(),
        GetOverrides()
      ]);
      variables = vars || {};
      overrides = ovr || {};
      mergeVariables();
      const count = Object.keys(mergedVariables).length;
      statusMessage = `Refreshed ${count} variable${count !== 1 ? 's' : ''}`;
      setTimeout(() => statusMessage = '', 3000);
    } catch (err) {
      console.error('Refresh error:', err);
      statusMessage = `Error: ${err}`;
    } finally {
      isRefreshing = false;
    }
  }

  async function openEnvFile() {
    try {
      await OpenEnvironmentFile();
    } catch (err) {
      console.error('Failed to open env file:', err);
    }
  }

  async function openOverridesFile() {
    try {
      await OpenOverridesFile();
    } catch (err) {
      console.error('Failed to open overrides file:', err);
    }
  }

  function startEdit(key: string, value: any) {
    editingKey = key;
    editValue = String(value);
  }

  function cancelEdit() {
    editingKey = null;
    editValue = '';
  }

  async function saveEdit() {
    if (editingKey === null) return;
    
    overrides[editingKey] = editValue;
    await saveOverridesToFile();
    editingKey = null;
    editValue = '';
    mergeVariables();
  }

  async function addOverride() {
    if (!newKey.trim()) return;
    
    overrides[newKey.trim()] = newValue;
    await saveOverridesToFile();
    newKey = '';
    newValue = '';
    mergeVariables();
  }

  async function removeOverride(key: string) {
    delete overrides[key];
    overrides = { ...overrides };
    await saveOverridesToFile();
    mergeVariables();
  }

  async function saveOverridesToFile() {
    isSaving = true;
    try {
      await SaveOverrides(overrides);
      statusMessage = 'Saved';
      setTimeout(() => statusMessage = '', 2000);
    } catch (err) {
      statusMessage = `Error saving: ${err}`;
    } finally {
      isSaving = false;
    }
  }

  function close() {
    editingKey = null;
    isOpen = false;
    dispatch('close');
  }

  function formatValue(value: any): string {
    if (typeof value === 'object') {
      return JSON.stringify(value, null, 2);
    }
    const str = String(value);
    if (str.length > 40) {
      return str.substring(0, 15) + '...' + str.substring(str.length - 10);
    }
    return str;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && editingKey !== null) {
      saveEdit();
    } else if (e.key === 'Escape') {
      if (editingKey !== null) {
        cancelEdit();
      } else {
        close();
      }
    }
  }
</script>

{#if isOpen}
  <div class="overlay" on:click={close} on:keydown={handleKeydown}>
    <div class="modal" on:click|stopPropagation>
      <div class="modal-header">
        <h2>Variables</h2>
        <button class="close-btn" on:click={close}>✕</button>
      </div>

      <div class="modal-content">
        <div class="toolbar">
          <div class="actions">
            <button class="btn" on:click={refreshVariables} disabled={isRefreshing}>
              {isRefreshing ? '⏳' : '⟳'} Refresh
            </button>
            <button class="btn secondary" on:click={openEnvFile} title="Edit environment variables">
              ✎ Env
            </button>
            <button class="btn secondary" on:click={openOverridesFile} title="Edit overrides file">
              ✎ Overrides
            </button>
          </div>
        </div>

        <div class="variables-list">
          {#each Object.entries(mergedVariables) as [key, { value, source }]}
            <div class="variable-row">
              <span class="var-key">{key}</span>
              
              {#if editingKey === key}
                <input 
                  class="var-input"
                  bind:value={editValue}
                  on:keydown={handleKeydown}
                  autofocus
                />
                <button class="icon-btn save" on:click={saveEdit} title="Save">✓</button>
                <button class="icon-btn cancel" on:click={cancelEdit} title="Cancel">✕</button>
              {:else}
                <span class="var-value" title={String(value)}>{formatValue(value)}</span>
                <span class="var-source" class:override={source === 'override'}>
                  {source === 'override' ? 'override' : 'env'}
                </span>
                {#if source === 'override'}
                  <button class="icon-btn edit" on:click={() => startEdit(key, value)} title="Edit">✎</button>
                  <button class="icon-btn delete" on:click={() => removeOverride(key)} title="Remove">✕</button>
                {/if}
              {/if}
            </div>
          {:else}
            <p class="empty">No variables loaded</p>
          {/each}
        </div>

        <!-- Add new override -->
        <div class="add-override">
          <input 
            class="add-input key"
            placeholder="Key"
            bind:value={newKey}
            on:keydown={(e) => e.key === 'Enter' && addOverride()}
          />
          <input 
            class="add-input value"
            placeholder="Value"
            bind:value={newValue}
            on:keydown={(e) => e.key === 'Enter' && addOverride()}
          />
          <button class="btn small" on:click={addOverride} disabled={!newKey.trim()}>
            + Add Override
          </button>
        </div>

        <p class="hint">
          Env variables use <code>$(command)</code> for dynamic tokens. 
          Overrides take precedence and can be edited inline.
        </p>
      </div>

      {#if statusMessage}
        <div class="status-bar" class:error={statusMessage.includes('Error')}>
          {statusMessage}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    width: 600px;
    max-width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .close-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: 18px;
    cursor: pointer;
    padding: 4px 8px;
  }

  .close-btn:hover {
    color: var(--text-primary);
  }

  .modal-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
  }

  .toolbar {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 12px;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .btn {
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    background: var(--accent-color);
    color: white;
    font-size: 12px;
    cursor: pointer;
    transition: opacity 0.15s;
  }

  .btn:hover:not(:disabled) {
    opacity: 0.9;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn.secondary {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
  }

  .btn.secondary:hover {
    background: var(--bg-hover);
  }

  .btn.small {
    padding: 4px 10px;
    font-size: 11px;
  }

  .variables-list {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    max-height: 280px;
    overflow-y: auto;
  }

  .variable-row {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-color);
    gap: 8px;
  }

  .variable-row:last-child {
    border-bottom: none;
  }

  .var-key {
    font-family: 'SF Mono', Monaco, monospace;
    font-size: 12px;
    color: var(--accent-color);
    min-width: 70px;
    font-weight: 500;
  }

  .var-value {
    font-family: 'SF Mono', Monaco, monospace;
    font-size: 12px;
    color: var(--text-secondary);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .var-input {
    flex: 1;
    padding: 4px 8px;
    border: 1px solid var(--accent-color);
    border-radius: 3px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, monospace;
    font-size: 12px;
  }

  .var-input:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .var-source {
    font-size: 9px;
    padding: 2px 5px;
    border-radius: 3px;
    background: var(--bg-primary);
    color: var(--text-muted);
    border: 1px solid var(--border-color);
    text-transform: uppercase;
  }

  .var-source.override {
    background: rgba(255, 193, 7, 0.15);
    color: #ffc107;
    border-color: rgba(255, 193, 7, 0.3);
  }

  .icon-btn {
    width: 24px;
    height: 24px;
    padding: 0;
    border: none;
    border-radius: 3px;
    background: transparent;
    color: var(--text-muted);
    font-size: 12px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .icon-btn:hover {
    background: var(--bg-hover);
  }

  .icon-btn.edit:hover {
    color: var(--accent-color);
  }

  .icon-btn.delete:hover {
    color: #f93e3e;
  }

  .icon-btn.save {
    color: #49cc90;
  }

  .icon-btn.cancel:hover {
    color: #f93e3e;
  }

  .add-override {
    display: flex;
    gap: 8px;
    margin-top: 12px;
    padding: 12px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
  }

  .add-input {
    padding: 6px 10px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-family: 'SF Mono', Monaco, monospace;
    font-size: 12px;
  }

  .add-input:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .add-input.key {
    width: 100px;
  }

  .add-input.value {
    flex: 1;
  }

  .empty {
    padding: 24px;
    text-align: center;
    color: var(--text-muted);
    font-size: 13px;
  }

  .hint {
    margin-top: 12px;
    font-size: 11px;
    color: var(--text-muted);
    line-height: 1.5;
  }

  .hint code {
    background: var(--bg-secondary);
    padding: 2px 5px;
    border-radius: 3px;
    font-family: 'SF Mono', Monaco, monospace;
    font-size: 10px;
  }

  .status-bar {
    padding: 10px 20px;
    background: rgba(73, 204, 144, 0.15);
    border-top: 1px solid var(--border-color);
    color: #49cc90;
    font-size: 12px;
    text-align: center;
  }

  .status-bar.error {
    background: rgba(249, 62, 62, 0.15);
    color: #f93e3e;
  }
</style>

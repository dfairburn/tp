<script lang="ts">
  import { onMount } from 'svelte';

  export let template = {
    name: '',
    url: '',
    method: 'GET',
    headers: [{ key: '', value: '' }],
    body: '',
    description: ''
  };
  export let onSave = (newTemplate) => {};
  export let onCancel = () => {};
  export let isNew = false;

  // Methods list for dropdown
  const methods = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];
  
  function addHeader() {
    template.headers.push({ key: '', value: '' });
  }

  function removeHeader(idx) {
    template.headers.splice(idx, 1);
  }

  function handleInput(e, idx, field) {
    template.headers[idx][field] = e.target.value;
  }

  function handleSave() {
    // Simple validation
    if (!template.name.trim() || !template.url.trim() || !template.method.trim()) return;
    onSave({ ...template });
  }
</script>

<div class="template-editor">
  <h2>{isNew ? 'New Template' : 'Edit Template'}{template.name && `: ${template.name}`}</h2>

  <div class="form-row">
    <label>Template Name</label>
    <input type="text" bind:value={template.name} placeholder="Name" autocomplete="off" />
  </div>
  <div class="form-row">
    <label>Description</label>
    <input type="text" bind:value={template.description} placeholder="Description (optional)" autocomplete="off" />
  </div>

  <div class="form-row">
    <label>HTTP Method</label>
    <select bind:value={template.method}>
      {#each methods as m}
        <option value={m}>{m}</option>
      {/each}
    </select>
  </div>

  <div class="form-row">
    <label>URL</label>
    <input type="text" bind:value={template.url} placeholder="https://..." autocomplete="off" />
  </div>

  <div class="form-row">
    <label>Headers</label>
    <div class="headers-list">
      {#each template.headers as hdr, idx}
        <div class="header-row">
          <input type="text" placeholder="Key" value={hdr.key} on:input={(e) => handleInput(e, idx, 'key')} />
          <input type="text" placeholder="Value" value={hdr.value} on:input={(e) => handleInput(e, idx, 'value')} />
          <button class="remove-btn" on:click={() => removeHeader(idx)} disabled={template.headers.length <= 1}>×</button>
        </div>
      {/each}
      <button class="add-btn" on:click={addHeader}>+ Add Header</button>
    </div>
  </div>

  <div class="form-row">
    <label>Body</label>
    <textarea bind:value={template.body} rows="8" spellcheck="false" placeholder="Body (JSON, GraphQL, etc.)"></textarea>
  </div>

  <div class="actions">
    <button class="cancel-btn" on:click={onCancel}>Cancel</button>
    <button class="save-btn" on:click={handleSave}>Save</button>
  </div>
</div>

<style>
  .template-editor {
    padding: 24px;
    background: var(--bg-secondary);
    border-radius: 12px;
    box-shadow: 0 2px 12px rgba(0,0,0,0.15);
    max-width: 540px;
    margin: auto;
  }
  h2 {
    margin-bottom: 18px;
    font-size: 22px;
    color: var(--accent-color);
  }
  .form-row {
    margin-bottom: 14px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  label {
    font-size: 13px;
    color: var(--text-secondary);
    font-weight: 600;
  }
  input, select, textarea {
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-size: 14px;
    padding: 8px;
    background: var(--bg-primary);
    color: var(--text-primary);
    width: 100%;
    box-sizing: border-box;
    outline: none;
  }
  input:focus, select:focus, textarea:focus {
    border-color: var(--accent-color);
  }
  .headers-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .header-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .add-btn, .remove-btn {
    background: var(--bg-hover);
    color: var(--text-secondary);
    border: none;
    padding: 4px 10px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
    transition: background 0.2s;
  }
  .add-btn:hover {
    background: var(--accent-color);
    color: #fff;
  }
  .remove-btn:disabled {
    opacity: 0.2;
    pointer-events: none;
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 20px;
  }
  .cancel-btn {
    background: var(--bg-hover);
    color: var(--text-secondary);
    border: none;
    padding: 8px 18px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
  }
  .save-btn {
    background: var(--accent-color);
    color: #fff;
    border: none;
    padding: 8px 18px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 16px;
  }
</style>

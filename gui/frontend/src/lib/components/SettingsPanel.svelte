<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { theme } from '../stores/app';
  import { formatValue } from './SettingsPanel';
  import styles from './SettingsPanel.module.css';
  import {
    GetVariables,
    RefreshVariables,
    GetEnvironmentFilePath,
    OpenEnvironmentFile,
    GetOverrides,
    OpenOverridesFile,
    SaveOverrides,
    GetConfig,
    SaveConfig,
  } from '../../../wailsjs/go/app/App';

  export let isOpen = false;

  const dispatch = createEventDispatcher();

  let settingsTab: 'variables' | 'config' = 'variables';

  // Variables tab state
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

  // Config tab state
  let editTemplatesDir = '';
  let editEnvFile = '';
  let configFilePath = '';
  let isSavingConfig = false;
  let configSaveError = '';

  $: if (isOpen) {
    loadData();
  }

  async function loadData() {
    try {
      const [vars, ovr, path, cfg] = await Promise.all([
        GetVariables(),
        GetOverrides(),
        GetEnvironmentFilePath(),
        GetConfig(),
      ]);
      variables = vars || {};
      overrides = ovr || {};
      envFilePath = path;
      mergeVariables();
      configFilePath = cfg.configPath || '';
      editTemplatesDir = cfg.templatesDirectoryPath || '';
      editEnvFile = cfg.environmentFile || '';
    } catch (err) {
      console.error('Failed to load variables:', err);
      statusMessage = `Error loading: ${err}`;
    }
  }

  async function saveConfig() {
    isSavingConfig = true;
    configSaveError = '';
    try {
      await SaveConfig(editEnvFile, editTemplatesDir);
      statusMessage = 'Config saved';
      setTimeout(() => statusMessage = '', 2000);
    } catch (err) {
      configSaveError = String(err);
    } finally {
      isSavingConfig = false;
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
  <div class={styles.overlay} on:click={close} on:keydown={handleKeydown}>
    <div class={styles.modal} on:click|stopPropagation>
      <div class={styles['modal-header']}>
        <h2>Settings</h2>
        <button class={styles['close-btn']} on:click={close}>✕</button>
      </div>

      <div class={styles['modal-tabs']}>
        <button class="{styles['tab-btn']} {settingsTab === 'variables' ? styles.active : ''}" on:click={() => settingsTab = 'variables'}>Variables</button>
        <button class="{styles['tab-btn']} {settingsTab === 'config' ? styles.active : ''}" on:click={() => settingsTab = 'config'}>Config</button>
      </div>

      <div class={styles['modal-content']}>

        {#if settingsTab === 'variables'}
          <div class={styles.toolbar}>
            <div class={styles.actions}>
              <button class={styles.btn} on:click={refreshVariables} disabled={isRefreshing}>
                {isRefreshing ? '⏳' : '⟳'} Refresh
              </button>
              <button class="{styles.btn} {styles.secondary}" on:click={openEnvFile} title="Edit environment variables">
                ✎ Env
              </button>
              <button class="{styles.btn} {styles.secondary}" on:click={openOverridesFile} title="Edit overrides file">
                ✎ Overrides
              </button>
            </div>
          </div>

          <div class={styles['variables-list']}>
            {#each Object.entries(mergedVariables) as [key, { value, source }]}
              <div class={styles['variable-row']}>
                <span class={styles['var-key']}>{key}</span>

                {#if editingKey === key}
                  <input
                    class={styles['var-input']}
                    bind:value={editValue}
                    on:keydown={handleKeydown}
                    autofocus
                  />
                  <button class="{styles['icon-btn']} {styles.save}" on:click={saveEdit} title="Save">✓</button>
                  <button class="{styles['icon-btn']} {styles.cancel}" on:click={cancelEdit} title="Cancel">✕</button>
                {:else}
                  <span class={styles['var-value']} title={String(value)}>{formatValue(value)}</span>
                  <span class="{styles['var-source']} {source === 'override' ? styles.override : ''}">
                    {source === 'override' ? 'override' : 'env'}
                  </span>
                  {#if source === 'override'}
                    <button class="{styles['icon-btn']} {styles.edit}" on:click={() => startEdit(key, value)} title="Edit">✎</button>
                    <button class="{styles['icon-btn']} {styles.delete}" on:click={() => removeOverride(key)} title="Remove">✕</button>
                  {/if}
                {/if}
              </div>
            {:else}
              <p class={styles.empty}>No variables loaded</p>
            {/each}
          </div>

          <div class={styles['add-override']}>
            <input
              class="{styles['add-input']} {styles.key}"
              placeholder="Key"
              bind:value={newKey}
              on:keydown={(e) => e.key === 'Enter' && addOverride()}
            />
            <input
              class="{styles['add-input']} {styles.value}"
              placeholder="Value"
              bind:value={newValue}
              on:keydown={(e) => e.key === 'Enter' && addOverride()}
            />
            <button class="{styles.btn} {styles.small}" on:click={addOverride} disabled={!newKey.trim()}>
              + Add Override
            </button>
          </div>

          <p class={styles.hint}>
            Env variables use <code>$(command)</code> for dynamic tokens.
            Overrides take precedence and can be edited inline.
          </p>

        {:else if settingsTab === 'config'}

          <div class={styles['config-section']}>
            <h3 class={styles['config-heading']}>Appearance</h3>
            <div class={styles['config-row']}>
              <span class={styles['config-label']}>Theme</span>
              <div class={styles['theme-toggle']}>
                <button class="{styles['theme-btn']} {$theme === 'dark' ? styles.active : ''}" on:click={() => theme.set('dark')}>
                  🌙 Dark
                </button>
                <button class="{styles['theme-btn']} {$theme === 'light' ? styles.active : ''}" on:click={() => theme.set('light')}>
                  ☀ Light
                </button>
              </div>
            </div>
          </div>

          <div class={styles['config-section']}>
            <h3 class={styles['config-heading']}>Paths</h3>
            <div class={styles['config-field']}>
              <label class={styles['config-label']}>Templates Directory</label>
              <input
                class={styles['config-input']}
                bind:value={editTemplatesDir}
                placeholder="~/.tp/templates"
                spellcheck="false"
                autocomplete="off"
              />
            </div>
            <div class={styles['config-field']}>
              <label class={styles['config-label']}>Environment File</label>
              <input
                class={styles['config-input']}
                bind:value={editEnvFile}
                placeholder="~/.tp/env.yml"
                spellcheck="false"
                autocomplete="off"
              />
            </div>
            <div class={styles['config-actions']}>
              <button class={styles.btn} on:click={saveConfig} disabled={isSavingConfig}>
                {isSavingConfig ? 'Saving…' : 'Save Config'}
              </button>
            </div>
            {#if configSaveError}
              <p class={styles['config-error']}>{configSaveError}</p>
            {/if}
          </div>

          <p class={styles.hint}>
            Config file: <code>{configFilePath || '~/.tp/config.yml'}</code><br/>
            Path changes take effect after restarting the app.
          </p>

        {/if}
      </div>

      {#if statusMessage}
        <div class="{styles['status-bar']} {statusMessage.includes('Error') ? styles.error : ''}">
          {statusMessage}
        </div>
      {/if}
    </div>
  </div>
{/if}


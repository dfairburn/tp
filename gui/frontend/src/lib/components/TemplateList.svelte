<script lang="ts">
  import { get } from 'svelte/store';
  import { templates, selectedTemplate, expandedDirs, searchQuery, filteredTemplates, toggleDirectory } from '../stores/app';
  import { GetTemplates, GetTemplate, CreateTemplate, DeleteTemplate, GetConfig, CreateFolder, RenameItem, MoveItem, SaveTemplate } from '../../../wailsjs/go/main/App';
  import type { main } from '../../../wailsjs/go/models';
  import { onMount } from 'svelte';

  let searchInput = '';
  let templatesRootDir = '';

  // Dialog state
  let showCreateDialog = false;
  let createDialogType: 'template' | 'folder' = 'template';
  let createDialogName = '';
  let createDialogParent = '';
  let createDialogError = '';

  // Rename dialog state
  let showRenameDialog = false;
  let renameTarget: main.TemplateItem | null = null;
  let renameDialogName = '';
  let renameDialogError = '';

  // Delete confirmation state
  let showDeleteConfirm = false;
  let deleteTarget: main.TemplateItem | null = null;

  // Hover state for action buttons
  let hoveredItem: string | null = null;

  // Drag and drop state
  let draggedItem: main.TemplateItem | null = null;
  let dropTargetPath: string | null = null;

  onMount(async () => {
    try {
      const config = await GetConfig();
      templatesRootDir = config.templatesDir;
    } catch (err) {
      console.error('Failed to get config:', err);
    }
    await loadTemplates();
  });

  async function loadTemplates() {
    try {
      const items = await GetTemplates();
      templates.set(items || []);

      let currentExpanded: Set<string>;
      expandedDirs.subscribe(v => currentExpanded = v)();

      if (currentExpanded.size === 0) {
        const dirs = new Set<string>();
        function collectDirs(items: main.TemplateItem[]) {
          for (const item of items) {
            if (item.isDir) {
              dirs.add(item.absolutePath);
              if (item.children) collectDirs(item.children);
            }
          }
        }
        collectDirs(items || []);
        expandedDirs.set(dirs);
      }
    } catch (err) {
      console.error('Failed to load templates:', err);
    }
  }

  function findByPath(items: main.TemplateItem[], path: string): main.TemplateItem | null {
    for (const item of items) {
      if (item.absolutePath === path) return item;
      if (item.children) {
        const found = findByPath(item.children, path);
        if (found) return found;
      }
    }
    return null;
  }

  function handleSearch(e: Event) {
    const target = e.target as HTMLInputElement;
    searchQuery.set(target.value);
  }

  function selectItem(item: main.TemplateItem) {
    if (item.isDir) {
      toggleDirectory(item.absolutePath);
    } else {
      selectedTemplate.set(item);
    }
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

  function isExpanded(path: string): boolean {
    return $expandedDirs.has(path);
  }

  // Create dialog
  function openCreateDialog(type: 'template' | 'folder', parentDir: string = '') {
    createDialogType = type;
    createDialogName = '';
    createDialogParent = parentDir;
    createDialogError = '';
    showCreateDialog = true;
    setTimeout(() => {
      const input = document.querySelector('.create-dialog input') as HTMLInputElement;
      input?.focus();
    }, 50);
  }

  async function handleCreate() {
    if (!createDialogName.trim()) {
      createDialogError = 'Name is required';
      return;
    }

    try {
      if (createDialogType === 'template') {
        const path = await CreateTemplate(createDialogName.trim(), createDialogParent);
        await loadTemplates();
        // Auto-select the newly created template
        const newItem = findByPath(get(templates), path);
        if (newItem) selectedTemplate.set(newItem);
      } else {
        await CreateFolder(createDialogName.trim(), createDialogParent);
        await loadTemplates();
      }
      showCreateDialog = false;
    } catch (err: any) {
      createDialogError = err?.message || 'Failed to create';
    }
  }

  function handleCreateKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      handleCreate();
    } else if (e.key === 'Escape') {
      showCreateDialog = false;
    }
  }

  // Delete
  function confirmDelete(item: main.TemplateItem) {
    deleteTarget = item;
    showDeleteConfirm = true;
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    try {
      await DeleteTemplate(deleteTarget.absolutePath);
      if ($selectedTemplate?.absolutePath === deleteTarget.absolutePath) {
        selectedTemplate.set(null);
      }
      await loadTemplates();
      showDeleteConfirm = false;
      deleteTarget = null;
    } catch (err) {
      console.error('Failed to delete:', err);
    }
  }

  // Edit: rename for folders, select for files (editing happens in RequestPanel)
  async function handleEdit(item: main.TemplateItem) {
    if (item.isDir) {
      openRenameDialog(item);
    } else {
      selectItem(item);
    }
  }

  // Rename dialog
  function openRenameDialog(item: main.TemplateItem) {
    renameTarget = item;
    renameDialogName = item.name;
    renameDialogError = '';
    showRenameDialog = true;
    setTimeout(() => {
      const input = document.querySelector('.rename-dialog input') as HTMLInputElement;
      if (input) {
        input.focus();
        input.select();
      }
    }, 50);
  }

  async function handleRename() {
    if (!renameTarget || !renameDialogName.trim()) {
      renameDialogError = 'Name is required';
      return;
    }

    if (renameDialogName.trim() === renameTarget.name) {
      showRenameDialog = false;
      return;
    }

    try {
      await RenameItem(renameTarget.absolutePath, renameDialogName.trim());
      await loadTemplates();
      showRenameDialog = false;
      renameTarget = null;
    } catch (err: any) {
      if (err?.message?.includes('exists')) {
        renameDialogError = 'An item with this name already exists';
      } else {
        renameDialogError = err?.message || 'Failed to rename';
      }
    }
  }

  function handleRenameKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      handleRename();
    } else if (e.key === 'Escape') {
      showRenameDialog = false;
    }
  }

  // Drag and drop
  let templatesContainer: HTMLElement;
  let scrollInterval: ReturnType<typeof setInterval> | null = null;
  const SCROLL_ZONE = 40;
  const SCROLL_SPEED = 8;

  function getTargetDirectory(item: main.TemplateItem): string {
    if (item.isDir) return item.absolutePath;
    const idx = item.absolutePath.lastIndexOf('/');
    if (idx === -1) return templatesRootDir || '';
    return item.absolutePath.substring(0, idx);
  }

  function handleDragStart(e: DragEvent, item: main.TemplateItem) {
    if (!e.dataTransfer) return;
    draggedItem = item;
    hoveredItem = null;
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', item.absolutePath);
  }

  function handleDragEnd() {
    draggedItem = null;
    dropTargetPath = null;
    stopAutoScroll();
  }

  function startAutoScroll(direction: 'up' | 'down') {
    if (scrollInterval) return;
    scrollInterval = setInterval(() => {
      if (!templatesContainer) return;
      templatesContainer.scrollTop += direction === 'up' ? -SCROLL_SPEED : SCROLL_SPEED;
    }, 16);
  }

  function stopAutoScroll() {
    if (scrollInterval) {
      clearInterval(scrollInterval);
      scrollInterval = null;
    }
  }

  function handleDragOverWithScroll(e: DragEvent) {
    if (!draggedItem || !templatesContainer) return;
    const rect = templatesContainer.getBoundingClientRect();
    const y = e.clientY - rect.top;
    if (y < SCROLL_ZONE) {
      startAutoScroll('up');
    } else if (y > rect.height - SCROLL_ZONE) {
      startAutoScroll('down');
    } else {
      stopAutoScroll();
    }
  }

  function handleDragOver(e: DragEvent, item: main.TemplateItem) {
    if (!draggedItem) return;
    const targetDir = getTargetDirectory(item);
    if (draggedItem.absolutePath === targetDir) return;
    if (targetDir.startsWith(draggedItem.absolutePath + '/')) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    if (dropTargetPath !== targetDir) dropTargetPath = targetDir;
  }

  function handleDragOverRoot(e: DragEvent) {
    if (!draggedItem) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    handleDragOverWithScroll(e);
    if (dropTargetPath !== templatesRootDir) dropTargetPath = templatesRootDir;
  }

  async function handleDrop(e: DragEvent, targetDir: string) {
    e.preventDefault();
    e.stopPropagation();
    if (!draggedItem) return;

    const lastSlash = draggedItem.absolutePath.lastIndexOf('/');
    let draggedParent = lastSlash === -1 ? templatesRootDir : draggedItem.absolutePath.substring(0, lastSlash);
    if (!targetDir) targetDir = templatesRootDir;
    if (!draggedParent) draggedParent = templatesRootDir;
    if (draggedParent === targetDir) {
      draggedItem = null;
      dropTargetPath = null;
      return;
    }

    const itemToMove = draggedItem;
    draggedItem = null;
    dropTargetPath = null;

    try {
      await MoveItem(itemToMove.absolutePath, targetDir);
      await loadTemplates();
    } catch (err: any) {
      console.error('Failed to move item:', err);
    }
  }
</script>

<div class="template-list">
  <div class="header-actions">
    <button class="header-btn" on:click={() => openCreateDialog('template')} title="New Template">
      + Template
    </button>
    <button class="header-btn" on:click={() => openCreateDialog('folder')} title="New Folder">
      + Folder
    </button>
  </div>

  <div class="search-box">
    <input
      type="text"
      placeholder="Search templates..."
      bind:value={searchInput}
      on:input={handleSearch}
    />
  </div>

  <div
    class="templates"
    class:is-dragging={draggedItem !== null}
    bind:this={templatesContainer}
    on:dragover={handleDragOverRoot}
    on:drop={(e) => handleDrop(e, templatesRootDir)}
    class:drop-target={dropTargetPath === templatesRootDir}
  >
    {#each $filteredTemplates as item}
      <div
        class="template-row"
        class:dragging={draggedItem?.absolutePath === item.absolutePath}
        class:drop-target={dropTargetPath === item.absolutePath}
        draggable="true"
        on:dragstart={(e) => handleDragStart(e, item)}
        on:dragend={handleDragEnd}
        on:dragover={(e) => handleDragOver(e, item)}
        on:drop={(e) => handleDrop(e, getTargetDirectory(item))}
        on:mouseenter={() => !draggedItem && (hoveredItem = item.absolutePath)}
        on:mouseleave={() => hoveredItem = null}
      >
        <button
          class="template-item"
          class:selected={$selectedTemplate?.absolutePath === item.absolutePath}
          class:directory={item.isDir}
          style="padding-left: {12 + item.depth * 16}px"
          on:click={() => selectItem(item)}
        >
          {#if item.isDir}
            <span class="icon">{isExpanded(item.absolutePath) ? '▼' : '▶'}</span>
            <span class="folder-icon">📁</span>
            <span class="name">{item.name}</span>
          {:else}
            <span class="method" style="color: {getMethodColor(item.method)}">{item.method || 'GET'}</span>
            <span class="name">{item.name}</span>
          {/if}
        </button>

        {#if hoveredItem === item.absolutePath}
          <div class="item-actions">
            {#if item.isDir}
              <button
                class="item-action-btn"
                on:click|stopPropagation={() => openCreateDialog('template', item.absolutePath)}
                title="New template in this folder"
              >+</button>
              <button
                class="item-action-btn"
                on:click|stopPropagation={() => handleEdit(item)}
                title="Rename"
              >✎</button>
            {/if}
            <button
              class="item-action-btn delete"
              on:click|stopPropagation={() => confirmDelete(item)}
              title="Delete"
            >×</button>
          </div>
        {/if}
      </div>
    {/each}

    {#if $filteredTemplates.length === 0}
      <div class="empty">No templates found</div>
    {/if}
  </div>

  <div class="actions">
    <button class="action-btn" on:click={loadTemplates} title="Reload templates">
      ⟳ Reload
    </button>
  </div>
</div>

<!-- Create Dialog -->
{#if showCreateDialog}
  <div class="dialog-overlay" on:click={() => showCreateDialog = false}>
    <div class="dialog create-dialog" on:click|stopPropagation>
      <h3>Create {createDialogType === 'template' ? 'Template' : 'Folder'}</h3>
      <input
        type="text"
        placeholder={createDialogType === 'template' ? 'Template name' : 'Folder name'}
        bind:value={createDialogName}
        on:keydown={handleCreateKeydown}
      />
      {#if createDialogError}
        <div class="error">{createDialogError}</div>
      {/if}
      <div class="dialog-actions">
        <button class="dialog-btn cancel" on:click={() => showCreateDialog = false}>Cancel</button>
        <button class="dialog-btn primary" on:click={handleCreate}>Create</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Dialog -->
{#if showDeleteConfirm && deleteTarget}
  <div class="dialog-overlay" on:click={() => showDeleteConfirm = false}>
    <div class="dialog delete-dialog" on:click|stopPropagation>
      <h3>Delete {deleteTarget.isDir ? 'Folder' : 'Template'}</h3>
      <p>Are you sure you want to delete <strong>{deleteTarget.name}</strong>?</p>
      {#if deleteTarget.isDir}
        <p class="warning">Warning: This will only delete the folder if it's empty.</p>
      {/if}
      <div class="dialog-actions">
        <button class="dialog-btn cancel" on:click={() => showDeleteConfirm = false}>Cancel</button>
        <button class="dialog-btn danger" on:click={handleDelete}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<!-- Rename Dialog -->
{#if showRenameDialog && renameTarget}
  <div class="dialog-overlay" on:click={() => showRenameDialog = false}>
    <div class="dialog rename-dialog" on:click|stopPropagation>
      <h3>Rename {renameTarget.isDir ? 'Folder' : 'Template'}</h3>
      <input
        type="text"
        placeholder="New name"
        bind:value={renameDialogName}
        on:keydown={handleRenameKeydown}
      />
      {#if renameDialogError}
        <div class="error">{renameDialogError}</div>
      {/if}
      <div class="dialog-actions">
        <button class="dialog-btn cancel" on:click={() => showRenameDialog = false}>Cancel</button>
        <button class="dialog-btn primary" on:click={handleRename}>Rename</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .template-list {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
  }

  .header-actions {
    display: flex;
    gap: 4px;
    padding: 8px;
    border-bottom: 1px solid var(--border-color);
  }

  .header-btn {
    flex: 1;
    padding: 6px 8px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 11px;
    cursor: pointer;
    transition: background 0.15s;
  }

  .header-btn:hover {
    background: var(--bg-hover);
    border-color: var(--accent-color);
  }

  .search-box {
    padding: 8px;
    border-bottom: 1px solid var(--border-color);
  }

  .search-box input {
    width: 100%;
    padding: 8px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 13px;
  }

  .search-box input:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .templates {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
  }

  .template-row {
    display: flex;
    align-items: center;
    position: relative;
  }

  .template-item {
    display: flex;
    align-items: center;
    gap: 8px;
    flex: 1;
    padding: 6px 12px;
    border: none;
    background: transparent;
    color: var(--text-primary);
    font-size: 13px;
    cursor: pointer;
    text-align: left;
    transition: background 0.15s;
  }

  .template-item:hover {
    background: var(--bg-hover);
  }

  .template-item.selected {
    background: var(--bg-selected);
  }

  .template-item.directory {
    color: var(--text-secondary);
  }

  .item-actions {
    display: flex;
    gap: 2px;
    padding-right: 8px;
    position: absolute;
    right: 0;
    background: linear-gradient(to right, transparent, var(--bg-secondary) 8px);
    padding-left: 16px;
  }

  .item-action-btn {
    width: 20px;
    height: 20px;
    border: none;
    border-radius: 3px;
    background: var(--bg-hover);
    color: var(--text-secondary);
    font-size: 14px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s;
  }

  .item-action-btn:hover {
    background: var(--accent-color);
    color: white;
  }

  .item-action-btn.delete:hover {
    background: #f93e3e;
    color: white;
  }

  .icon {
    font-size: 10px;
    width: 12px;
    color: var(--text-muted);
  }

  .folder-icon {
    font-size: 14px;
  }

  .method {
    font-size: 10px;
    font-weight: 600;
    min-width: 45px;
    text-transform: uppercase;
  }

  .name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty {
    padding: 20px;
    text-align: center;
    color: var(--text-muted);
  }

  .actions {
    padding: 8px;
    border-top: 1px solid var(--border-color);
    display: flex;
    gap: 8px;
  }

  .action-btn {
    flex: 1;
    padding: 6px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 12px;
    cursor: pointer;
    transition: background 0.15s;
  }

  .action-btn:hover {
    background: var(--bg-hover);
  }

  /* Dialog styles */
  .dialog-overlay {
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

  .dialog {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 20px;
    min-width: 300px;
    max-width: 400px;
  }

  .dialog h3 {
    margin: 0 0 16px 0;
    font-size: 16px;
    color: var(--text-primary);
  }

  .dialog input {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 14px;
    margin-bottom: 12px;
    box-sizing: border-box;
  }

  .dialog input:focus {
    outline: none;
    border-color: var(--accent-color);
  }

  .dialog p {
    margin: 0 0 12px 0;
    color: var(--text-secondary);
    font-size: 14px;
  }

  .dialog .warning {
    color: #fca130;
    font-size: 12px;
  }

  .dialog .error {
    color: #f93e3e;
    font-size: 12px;
    margin-bottom: 12px;
  }

  .dialog-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  .dialog-btn {
    padding: 8px 16px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-size: 13px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .dialog-btn.cancel {
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .dialog-btn.cancel:hover {
    background: var(--bg-hover);
  }

  .dialog-btn.primary {
    background: var(--accent-color);
    border-color: var(--accent-color);
    color: white;
  }

  .dialog-btn.primary:hover {
    filter: brightness(1.1);
  }

  .dialog-btn.danger {
    background: #f93e3e;
    border-color: #f93e3e;
    color: white;
  }

  .dialog-btn.danger:hover {
    filter: brightness(1.1);
  }

  /* Drag and drop styles */
  .template-row.dragging {
    opacity: 0.4;
  }

  .template-row.drop-target {
    background: color-mix(in srgb, var(--accent-color) 30%, transparent);
    border-radius: 4px;
    outline: 2px solid var(--accent-color);
    outline-offset: -2px;
  }

  .template-row.drop-target .template-item {
    background: transparent;
  }

  .templates.drop-target {
    background: var(--bg-hover);
    outline: 2px dashed var(--accent-color);
    outline-offset: -4px;
    border-radius: 4px;
  }

  .templates.is-dragging .item-actions {
    display: none;
  }
</style>

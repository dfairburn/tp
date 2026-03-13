<script lang="ts">
  import { get } from 'svelte/store';
  import { templates, selectedTemplate, expandedDirs, searchQuery, filteredTemplates, toggleDirectory } from '../stores/app';
  import { GetTemplates, GetTemplate, CreateTemplate, DeleteTemplate, GetConfig, CreateFolder, RenameItem, MoveItem, SaveTemplate } from '../../../wailsjs/go/app/App';
  import type { app } from '../../../wailsjs/go/models';
  import { onMount } from 'svelte';
  import { getMethodColor } from '../utils';
  import styles from './TemplateList.module.css';

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
  let renameTarget: app.TemplateItem | null = null;
  let renameDialogName = '';
  let renameDialogError = '';

  // Delete confirmation state
  let showDeleteConfirm = false;
  let deleteTarget: app.TemplateItem | null = null;

  // Hover state for action buttons
  let hoveredItem: string | null = null;

  // Drag and drop state
  let draggedItem: app.TemplateItem | null = null;
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
        function collectDirs(items: app.TemplateItem[]) {
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

  function findByPath(items: app.TemplateItem[], path: string): app.TemplateItem | null {
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

  function selectItem(item: app.TemplateItem) {
    if (item.isDir) {
      toggleDirectory(item.absolutePath);
    } else {
      selectedTemplate.set(item);
    }
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
  function confirmDelete(item: app.TemplateItem) {
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
  async function handleEdit(item: app.TemplateItem) {
    if (item.isDir) {
      openRenameDialog(item);
    } else {
      selectItem(item);
    }
  }

  // Rename dialog
  function openRenameDialog(item: app.TemplateItem) {
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

  function getTargetDirectory(item: app.TemplateItem): string {
    if (item.isDir) return item.absolutePath;
    const idx = item.absolutePath.lastIndexOf('/');
    if (idx === -1) return templatesRootDir || '';
    return item.absolutePath.substring(0, idx);
  }

  function handleDragStart(e: DragEvent, item: app.TemplateItem) {
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

  function handleDragOver(e: DragEvent, item: app.TemplateItem) {
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

<div class={styles['template-list']}>
  <div class={styles['header-actions']}>
    <button class={styles['header-btn']} on:click={() => openCreateDialog('template')} title="New Template">
      + Template
    </button>
    <button class={styles['header-btn']} on:click={() => openCreateDialog('folder')} title="New Folder">
      + Folder
    </button>
  </div>

  <div class={styles['search-box']}>
    <input
      type="text"
      placeholder="Search templates..."
      bind:value={searchInput}
      on:input={handleSearch}
    />
  </div>

  <div
    class="{styles.templates} {draggedItem !== null ? styles['is-dragging'] : ''} {dropTargetPath === templatesRootDir ? styles['drop-target'] : ''}"
    bind:this={templatesContainer}
    on:dragover={handleDragOverRoot}
    on:drop={(e) => handleDrop(e, templatesRootDir)}
  >
    {#each $filteredTemplates as item}
      <div
        class="{styles['template-row']} {draggedItem?.absolutePath === item.absolutePath ? styles.dragging : ''} {dropTargetPath === item.absolutePath ? styles['drop-target'] : ''}"
        draggable="true"
        on:dragstart={(e) => handleDragStart(e, item)}
        on:dragend={handleDragEnd}
        on:dragover={(e) => handleDragOver(e, item)}
        on:drop={(e) => handleDrop(e, getTargetDirectory(item))}
        on:mouseenter={() => !draggedItem && (hoveredItem = item.absolutePath)}
        on:mouseleave={() => hoveredItem = null}
      >
        <button
          class="{styles['template-item']} {$selectedTemplate?.absolutePath === item.absolutePath ? styles.selected : ''} {item.isDir ? styles.directory : ''}"
          style="padding-left: {12 + item.depth * 16}px"
          on:click={() => selectItem(item)}
        >
          {#if item.isDir}
            <span class={styles.icon}>{isExpanded(item.absolutePath) ? '▼' : '▶'}</span>
            <span class={styles['folder-icon']}>📁</span>
            <span class={styles.name}>{item.name}</span>
          {:else}
            <span class={styles.method} style="color: {getMethodColor(item.method)}">{item.method || 'GET'}</span>
            <span class={styles.name}>{item.name}</span>
          {/if}
        </button>

        {#if hoveredItem === item.absolutePath}
          <div class={styles['item-actions']}>
            {#if item.isDir}
              <button
                class={styles['item-action-btn']}
                on:click|stopPropagation={() => openCreateDialog('template', item.absolutePath)}
                title="New template in this folder"
              >+</button>
              <button
                class={styles['item-action-btn']}
                on:click|stopPropagation={() => handleEdit(item)}
                title="Rename"
              >✎</button>
            {/if}
            <button
              class="{styles['item-action-btn']} {styles.delete}"
              on:click|stopPropagation={() => confirmDelete(item)}
              title="Delete"
            >×</button>
          </div>
        {/if}
      </div>
    {/each}

    {#if $filteredTemplates.length === 0}
      <div class={styles.empty}>No templates found</div>
    {/if}
  </div>

  <div class={styles.actions}>
    <button class={styles['action-btn']} on:click={loadTemplates} title="Reload templates">
      ⟳ Reload
    </button>
  </div>
</div>

<!-- Create Dialog -->
{#if showCreateDialog}
  <div class={styles['dialog-overlay']} on:click={() => showCreateDialog = false}>
    <div class={styles.dialog} on:click|stopPropagation>
      <h3>Create {createDialogType === 'template' ? 'Template' : 'Folder'}</h3>
      <input
        type="text"
        placeholder={createDialogType === 'template' ? 'Template name' : 'Folder name'}
        bind:value={createDialogName}
        on:keydown={handleCreateKeydown}
      />
      {#if createDialogError}
        <div class={styles.error}>{createDialogError}</div>
      {/if}
      <div class={styles['dialog-actions']}>
        <button class="{styles['dialog-btn']} {styles.cancel}" on:click={() => showCreateDialog = false}>Cancel</button>
        <button class="{styles['dialog-btn']} {styles.primary}" on:click={handleCreate}>Create</button>
      </div>
    </div>
  </div>
{/if}

<!-- Delete Confirmation Dialog -->
{#if showDeleteConfirm && deleteTarget}
  <div class={styles['dialog-overlay']} on:click={() => showDeleteConfirm = false}>
    <div class={styles.dialog} on:click|stopPropagation>
      <h3>Delete {deleteTarget.isDir ? 'Folder' : 'Template'}</h3>
      <p>Are you sure you want to delete <strong>{deleteTarget.name}</strong>?</p>
      {#if deleteTarget.isDir}
        <p class={styles.warning}>Warning: This will only delete the folder if it's empty.</p>
      {/if}
      <div class={styles['dialog-actions']}>
        <button class="{styles['dialog-btn']} {styles.cancel}" on:click={() => showDeleteConfirm = false}>Cancel</button>
        <button class="{styles['dialog-btn']} {styles.danger}" on:click={handleDelete}>Delete</button>
      </div>
    </div>
  </div>
{/if}

<!-- Rename Dialog -->
{#if showRenameDialog && renameTarget}
  <div class={styles['dialog-overlay']} on:click={() => showRenameDialog = false}>
    <div class={styles.dialog} on:click|stopPropagation>
      <h3>Rename {renameTarget.isDir ? 'Folder' : 'Template'}</h3>
      <input
        type="text"
        placeholder="New name"
        bind:value={renameDialogName}
        on:keydown={handleRenameKeydown}
      />
      {#if renameDialogError}
        <div class={styles.error}>{renameDialogError}</div>
      {/if}
      <div class={styles['dialog-actions']}>
        <button class="{styles['dialog-btn']} {styles.cancel}" on:click={() => showRenameDialog = false}>Cancel</button>
        <button class="{styles['dialog-btn']} {styles.primary}" on:click={handleRename}>Rename</button>
      </div>
    </div>
  </div>
{/if}


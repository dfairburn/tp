<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import styles from './ParamRow.module.css';

  export let varName: string;
  export let value: string;
  export let placeholder: string = 'Value...';
  export let description: string | undefined = undefined;

  const dispatch = createEventDispatcher<{ change: string }>();

  function handleInput(e: Event) {
    dispatch('change', (e.target as HTMLInputElement).value);
  }
</script>

<code class={styles['pgrid-key']}>{'{{.'}{varName}{'}}'}</code>
<input
  type="text"
  class={styles['pgrid-value']}
  {value}
  on:input={handleInput}
  {placeholder}
  autocomplete="off"
  spellcheck="false"
/>
{#if description !== undefined}
  <span class={styles['pgrid-desc']}>{description}</span>
{/if}

<script lang="ts">
  import { isDarkMode } from '../../../../src/stores';
  import { createEventDispatcher } from 'svelte';
  
  export let show: boolean = false;
  export let title: string = '';
  export let showCloseButton: boolean = true;
  
  const dispatch = createEventDispatcher<{ close: void }>();
  
  function close(): void {
    show = false;
    dispatch('close');
  }
  
  function handleClickOutside(event: MouseEvent): void {
    if ((event.target as Element)?.classList.contains('modal-overlay')) {
      close();
    }
  }
  
  function handleKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      close();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
  <div class="modal-overlay" role="button" tabindex="0" on:click={handleClickOutside} on:keydown={handleKeydown}>
    <div role="dialog" aria-modal="true" aria-labelledby={title ? 'modal-title' : null}
        class="modal-content" 
        class:bg-gray-100={!$isDarkMode}
        class:bg-[#242423]={$isDarkMode}

        style="border: 1px solid var(--card-border);"
    >
      {#if title || showCloseButton}
        <header class="modal-header text-gray-600 dark:text-gray-200">
          {#if title}
            <h2 id="modal-title">{title}</h2>
          {/if}
          {#if showCloseButton}
            <button on:click={close}>
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
              </svg>
            </button>
          {/if}
        </header>
      {/if}
      
      <div class="modal-body">
        <slot />
      </div>
      
      <footer class="modal-footer">
        <slot name="footer" />
      </footer>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .modal-content {
    border-radius: 8px;
    max-width: 500px;
    width: 90%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid var(--card-border);;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .modal-body {
    padding: 1.5rem;
  }

  .modal-footer {
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--card-border);
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }

  .modal-footer:empty {
    display: none;
  }
</style>
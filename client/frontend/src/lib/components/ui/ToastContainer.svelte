<script lang="ts">
  import { toasts, removeToast } from '../../../stores/toastStore';
  import type { Toast } from '../../../stores/toastStore';
  import { fly, fade } from 'svelte/transition';
</script>

<style>
  .toast {
    padding: 1rem;
    margin: 0.5rem 0;
    border-radius: 0.375rem;
    color: white;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    cursor: pointer;
  }
  .success { background-color: #48bb78; }
  .info    { background-color: #4299e1; }
  .warning { background-color: #ed8936; }
  .error   { background-color: #f56565; }
</style>

<!-- fixed container in top‑right -->
<div class="fixed top-4 right-4 flex flex-col items-end pointer-events-none">
  {#each $toasts as toast (toast.id)}
    <button
      class="toast {toast.type} pointer-events-auto"
      in:fly={{ x: 200, duration: 300 }}
      out:fade={{ duration: 200 }}
      on:click={() => removeToast(toast.id)}
    >
      {toast.message}
    </button>
  {/each}
</div>

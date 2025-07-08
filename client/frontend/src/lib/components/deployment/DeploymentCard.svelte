<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { parseDuration, parseDate } from '../../../utils';
  import type { Deployment } from '../../../types';

  export let deployment: Deployment;
  export let isSelected: boolean = false;

  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('select');
  }
</script>

<button 
  class="card w-full p-4 rounded-lg border-l-4"
  class:border-green-500={deployment.status === 'success'}
  class:border-red-500={deployment.status === 'failed'}
  class:border-yellow-500={deployment.status === 'pending'}
  class:bg-gray-100={isSelected}
  class:dark:bg-opacity-10={isSelected}
  on:click={handleClick}
>
   <div class="flex items-center justify-between mb-2 min-w-0">
      <div class="flex items-center space-x-3 min-w-0 flex-1">
          <span class="text-xs px-2 py-1 rounded flex-shrink-0"
              class:bg-green-100={deployment.status === 'success'}
              class:text-green-800={deployment.status === 'success'}
              class:bg-red-100={deployment.status === 'failed'}
              class:text-red-800={deployment.status === 'failed'}
              class:bg-yellow-100={deployment.status === 'pending'}
              class:text-yellow-800={deployment.status === 'pending'}
          >{(deployment.status)?.toUpperCase()}</span>
          <span class="font-mono text-sm truncate">{deployment.id}</span>
      </div>
      <span class="text-sm text-gray-500 flex-shrink-0 ml-2">{parseDuration(deployment.duration)}</span>
  </div>
  <div class="flex items-center justify-between text-sm min-w-0">
      <div class="flex items-center space-x-2 min-w-0 flex-1">
          <span class="flex-shrink-0 flex gap-1">
            <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-git-branch"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 18m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M7 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M17 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M7 8l0 8" /><path d="M9 18h6a2 2 0 0 0 2 -2v-5" /><path d="M14 14l3 -3l3 3" /></svg>
            {deployment.branch}</span>
          <span class="font-mono truncate">{deployment.commitHash}</span>
          <span class="truncate">{deployment.message}</span>
      </div>
      <div class="flex items-center space-x-2 flex-shrink-0 ml-2">
          <span class="truncate">{parseDate(deployment.createdAt, deployment.user)}</span>
          <div class="w-6 h-6 bg-orange-500 rounded-full flex-shrink-0"></div>
      </div>
  </div>
</button>
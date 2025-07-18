<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { gitIcon, formatProjectDate, handleKeydown } from '../../../utils';
  import type { Project } from '../../../types';
  import { sidebarWidth } from '../../../../src/stores';
  import { SIDEBAR_WIDTH_DOCKED } from '../../../../src/constants';

  export let project: Project;
  export let isSelected: boolean;

  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('select');
  }
</script>

<div role="button"
  tabindex="0"
  aria-pressed={isSelected}
  class="card rounded-lg cursor-pointer transition-all duration-200 text-left relative"
  class:bg-gray-100={isSelected}
  class:dark:bg-opacity-10={isSelected}
  class:p-6={$sidebarWidth !== SIDEBAR_WIDTH_DOCKED}
  class:h-24={$sidebarWidth === SIDEBAR_WIDTH_DOCKED}
  on:click={handleClick}
  on:keydown={(e) => handleKeydown(e, handleClick)}
>
  {#if $sidebarWidth !== SIDEBAR_WIDTH_DOCKED}
    <button class="absolute top-4 right-4 dark:text-gray-400 text-gray-600 font-medium hover:text-white" tabindex="-1">⋯</button>
    <div class="pr-8 mb-4">
      <div class="flex items-center space-x-3">
        <img src={project.logo} alt={project.name} class="w-10 h-10 rounded" />
        <div class="min-w-0 flex-1">
          <h3 class="font-semibold truncate">{project.name}</h3>
          <p class="text-sm dark:text-gray-400 font-medium text-gray-600 truncate">{project.description}</p>
        </div>
      </div>
    </div>
    
    <div class="flex items-center">
      <div class="flex items-center space-x-2 min-w-0">
        {@html gitIcon(project.provider)}
        <span class="text-sm dark:text-gray-400 text-gray-600 font-medium truncate">{project.url}</span>
      </div>
      <div class="flex ml-auto text-sm dark:text-gray-400 font-medium text-gray-600 flex-shrink-0">
        {formatProjectDate(project.date)}
      </div>
    </div>
  {:else}
    <div class="absolute inset-0 flex items-center justify-center overflow-hidden">
      <p class="transform -rotate-90 whitespace-nowrap text-sm font-medium">
        {project.name}
      </p>
    </div>
  {/if}
</div>
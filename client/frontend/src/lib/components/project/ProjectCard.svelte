<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { gitIcon, formatProjectDate } from '../../../utils';
  import type { Project } from '../../../types';

  export let project: Project;
  export let isSelected: boolean = false;

  const dispatch = createEventDispatcher();

  function handleClick() {
    dispatch('select');
  }
</script>

<button type="button"
  class="card p-6 rounded-lg cursor-pointer transition-all duration-200 text-left relative"  
  class:bg-gray-100={isSelected}
  class:dark:bg-opacity-10={isSelected}
  on:click={handleClick}
  tabindex="0"
  aria-pressed={isSelected}
>
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
</button>

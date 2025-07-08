<script lang="ts">
  import { projects, selectedProject, appState } from '../../../stores';
  import { gitIcon, formatProjectDate } from '../../../utils';
  
  function selectProject(project: any) {
    selectedProject.set(project);
    appState.update(state => ({ ...state, selectedProjectId: project.name }));
  }
</script>

<div class="space-y-3">
  {#each $projects as project}
    <button type="button"
      class="card p-4 rounded-lg cursor-pointer transition-all duration-200 text-left w-full relative"  
      class:bg-gray-100={$appState.selectedProjectId === project.name}
      class:dark:bg-opacity-10={$appState.selectedProjectId === project.name}
      on:click={() => selectProject(project)}
      tabindex="0"
      aria-pressed={$appState.selectedProjectId === project.name}
    >
      <button class="absolute top-4 right-4 dark:text-gray-400 text-gray-600 font-medium hover:text-white flex-shrink-0" tabindex="-1">⋯</button>
      <div class="pr-8">
          <div class="flex items-center space-x-3 min-w-0 flex-1">
              <img src={project.icon} alt={project.name} class="w-8 h-8 rounded flex-shrink-0" />
              <div class="min-w-0 flex-1">
                  <h3 class="font-semibold truncate">{project.name}</h3>
                  <p class="text-sm dark:text-gray-400 font-medium text-gray-600 truncate">{project.description}</p>
              </div>
          </div>
          <div class="flex items-center space-x-4 min-w-0 flex-1 mt-2">
              <div class="flex items-center space-x-2 min-w-0">
                  {@html gitIcon(project.provider)}
                  <span class="text-sm dark:text-gray-400 font-medium text-gray-600 truncate">{project.url}</span>
              </div>
              <div class="flex ml-auto text-sm dark:text-gray-400 font-medium text-gray-600 flex-shrink-0">
                  {formatProjectDate(project.date)}
              </div>
          </div>
      </div>
    </button>
  {/each}

  <!-- Add New Project Row -->
  <button type="button" 
    class="card p-4 rounded-lg cursor-pointer transition-all duration-200 text-left w-full border-dashed"
    on:click={() => {}}
    tabindex="0"
  >
      <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3 min-w-0 flex-1">
              <div class="w-8 h-8 rounded bg-gray-700 flex items-center justify-center flex-shrink-0">
                  <span class="text-white text-lg">+</span>
              </div>
              <div class="min-w-0 flex-1">
                  <h3 class="font-semibold text-gray-400">Create a new project</h3>
                  <p class="text-sm dark:text-gray-400 font-medium text-gray-600">Start building something amazing</p>
              </div>
          </div>
          <div class="flex items-center space-x-4 min-w-0 flex-1">
              <div class="flex items-center space-x-2 min-w-0">
                  <span class="text-sm dark:text-gray-400 font-medium text-gray-600">Click to get started</span>
              </div>
          </div>
      </div>
  </button>
</div>

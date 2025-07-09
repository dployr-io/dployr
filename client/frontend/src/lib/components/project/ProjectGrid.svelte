<script lang="ts">
  import { projects, selectedProject, appState } from '../../../stores';
  import ProjectCard from './ProjectCard.svelte';
  
  export let gridCols: number;
  export let maxGridCols: number;

  function selectProject(project: any) {
    selectedProject.set(project);
    appState.update(state => ({ ...state, selectedProjectId: project.name }));
  }
</script>

<div class="grid gap-6" style="grid-template-columns: repeat({Math.min(gridCols, maxGridCols)}, minmax(300px, 1fr));">
  {#each $projects as project}
    <ProjectCard 
      {project} 
      isSelected={$appState.selectedProjectId === project.name}
      on:select={() => selectProject(project)}
    />
  {/each}

  <!-- Add New Project Card -->
  <button type="button" 
    class="card p-6 rounded-lg cursor-pointer transition-all duration-200 text-left border-dashed"
    on:click={() => {}}
    tabindex="0"
  >
      <div class="flex items-start justify-between mb-4">
          <div class="flex items-center space-x-3">
              <div class="w-10 h-10 rounded dark:bg-gray-700 bg-gray-500 flex items-center justify-center">
                  <span class="text-white text-xl">+</span>
              </div>
              <div class="min-w-0 flex-1">
                  <h3 class="font-semibold dark:text-gray-400 text-gray-600">Create a new project</h3>
                  <p class="text-sm dark:text-gray-400 font-medium text-gray-600">Start building something amazing</p>
              </div>
          </div>
      </div>
      <div class="flex items-center">
          <div class="flex items-center space-x-2 min-w-0">
              <span class="text-sm dark:text-gray-400 text-gray-600 font-medium">Click to get started</span>
          </div>
      </div>
  </button>
</div>

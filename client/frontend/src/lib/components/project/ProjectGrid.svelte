<script lang="ts">
  import { SIDEBAR_WIDTH_DOCKED } from '../../../../src/constants';
  import { projects, selectedProject, appState, sidebarWidth, showNewProjectPopup } from '../../../stores';
  import ProjectCard from './ProjectCard.svelte';


  function selectProject(project: any) {
    selectedProject.set(project);
    appState.update(state => ({ ...state, selectedProjectId: project.id }));
  }

  function handleSideBarDockToggle() {
    $sidebarWidth !== SIDEBAR_WIDTH_DOCKED ?
    sidebarWidth.set(SIDEBAR_WIDTH_DOCKED):
    sidebarWidth.set(640);
  } 

  function handleNewProject() {
    showNewProjectPopup.set(true);
  }
</script>
<div class="flex flex-col gap-4">
  <div class="flex w-full justify-end items-center">
    <button 
      on:click={handleSideBarDockToggle}
    >
      {#if $sidebarWidth === SIDEBAR_WIDTH_DOCKED}
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="m5.25 4.5 7.5 7.5-7.5 7.5m6-15 7.5 7.5-7.5 7.5" />
        </svg>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="m18.75 4.5-7.5 7.5 7.5 7.5m-6-15L5.25 12l7.5 7.5" />
        </svg>
      {/if}
    </button>
  </div>
  <div class="flex flex-col gap-6 items-center">
    {#if $sidebarWidth === SIDEBAR_WIDTH_DOCKED}
      <button on:click={handleNewProject}>
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v6m3-3H9m12 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
      </button>
    {/if}
    {#each $projects as project}
      <ProjectCard 
        {project} 
        isSelected={$selectedProject?.id === project.id}
        on:select={() => selectProject(project)}
      />
    {/each}
  
    {#if $sidebarWidth !== SIDEBAR_WIDTH_DOCKED}
      <!-- Add New Project Card -->
      <button type="button" 
        class="card p-6 w-full rounded-lg cursor-pointer transition-all duration-200 text-left border-dashed"
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
    {/if}
  </div>
</div>

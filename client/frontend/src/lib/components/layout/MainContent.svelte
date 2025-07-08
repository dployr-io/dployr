<script lang="ts">
  import { selectedProject, appState, viewMode } from '../../../stores';
  import ProjectGrid from '../project/ProjectGrid.svelte';
  import ProjectList from '../project/ProjectList.svelte';
  
  export let sidebarWidth: number;

  let selectedSection = 'Deployments';

  function selectSection(section: string) {
    selectedSection = section;
    appState.update(state => ({ ...state, selectedSection: section }));
  }

  // Calculate grid columns based on available width
  $: mainContentWidth = typeof window !== 'undefined' ? window.innerWidth - sidebarWidth - 48 : 800; // 48px for padding
  $: gridCols = Math.max(1, Math.floor(mainContentWidth / 350)); // 350px min card width
</script>

<!-- Project Header with Navigation -->
<div class="flex items-center justify-between mb-6">
    <h2 class="text-2xl font-bold">{$selectedProject?.name || 'Select a Project'}</h2>
    
    <!-- Section Navigation Icons -->
    <div class="flex items-center space-x-2">
        <!-- Deployments -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Deployments'}
            class:dark:bg-gray-200={selectedSection === 'Deployments'}
            class:text-gray-500={selectedSection === 'Deployments'}
            class:dark:text-gray-500={selectedSection === 'Deployments'}
            class:dark:text-gray-400={selectedSection !== 'Deployments'}
            class:hover:bg-gray-100={selectedSection !== 'Deployments'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Deployments'}

            on:click={() => selectSection('Deployments')}
            title="Deployments"
        >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3" />
            </svg>
        </button>

        <!-- Logs -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Logs'}
            class:dark:bg-gray-200={selectedSection === 'Logs'}
            class:text-gray-500={selectedSection === 'Logs'}
            class:dark:text-gray-500={selectedSection === 'Logs'}
            class:dark:text-gray-400={selectedSection !== 'Logs'}
            class:hover:bg-gray-100={selectedSection !== 'Logs'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Logs'}
            on:click={() => selectSection('Logs')}
            title="Logs"
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
          </svg>
        </button>

        <!-- Resources -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Resources'}
            class:dark:bg-gray-200={selectedSection === 'Resources'}
            class:text-gray-500={selectedSection === 'Resources'}
            class:dark:text-gray-500={selectedSection === 'Resources'}
            class:dark:text-gray-400={selectedSection !== 'Resources'}
            class:hover:bg-gray-100={selectedSection !== 'Resources'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Resources'}
            on:click={() => selectSection('Resources')}
            title="Resources"
        >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 0 1-3-3m3 3a3 3 0 1 0 0 6h13.5a3 3 0 1 0 0-6m-16.5-3a3 3 0 0 1 3-3h13.5a3 3 0 0 1 3 3m-19.5 0a4.5 4.5 0 0 1 .9-2.7L5.737 5.1a3.375 3.375 0 0 1 2.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 0 1 .9 2.7m0 0a3 3 0 0 1-3 3m0 3h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Zm-3 6h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Z" />
            </svg>
        </button>

        <!-- Domains -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Domains'}
            class:dark:bg-gray-200={selectedSection === 'Domains'}
            class:text-gray-500={selectedSection === 'Domains'}
            class:dark:text-gray-500={selectedSection === 'Domains'}
            class:dark:text-gray-400={selectedSection !== 'Domains'}
            class:hover:bg-gray-100={selectedSection !== 'Domains'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Domains'}
            on:click={() => selectSection('Domains')}
            title="Domains"
        >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
        </svg>
        </button>

        <!-- Settings -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Settings'}
            class:dark:bg-gray-200={selectedSection === 'Settings'}
            class:text-gray-500={selectedSection === 'Settings'}
            class:dark:text-gray-500={selectedSection === 'Settings'}
            class:dark:text-gray-400={selectedSection !== 'Settings'}
            class:hover:bg-gray-100={selectedSection !== 'Settings'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Settings'}
            on:click={() => selectSection('Settings')}
            title="Settings"
        >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.343 3.94c.09-.542.56-.94 1.11-.94h1.093c.55 0 1.02.398 1.11.94l.149.894c.07.424.384.764.78.93.398.164.855.142 1.205-.108l.737-.527a1.125 1.125 0 0 1 1.45.12l.773.774c.39.389.44 1.002.12 1.45l-.527.737c-.25.35-.272.806-.107 1.204.165.397.505.71.93.78l.893.15c.543.09.94.559.94 1.109v1.094c0 .55-.397 1.02-.94 1.11l-.894.149c-.424.07-.764.383-.929.78-.165.398-.143.854.107 1.204l.527.738c.32.447.269 1.06-.12 1.45l-.774.773a1.125 1.125 0 0 1-1.449.12l-.738-.527c-.35-.25-.806-.272-1.203-.107-.398.165-.71.505-.781.929l-.149.894c-.09.542-.56.94-1.11.94h-1.094c-.55 0-1.019-.398-1.11-.94l-.148-.894c-.071-.424-.384-.764-.781-.93-.398-.164-.854-.142-1.204.108l-.738.527c-.447.32-1.06.269-1.45-.12l-.773-.774a1.125 1.125 0 0 1-.12-1.45l.527-.737c.25-.35.272-.806.108-1.204-.165-.397-.506-.71-.93-.78l-.894-.15c-.542-.09-.94-.56-.94-1.109v-1.094c0-.55.398-1.02.94-1.11l.894-.149c.424-.07.765-.383.93-.78.165-.398.143-.854-.108-1.204l-.526-.738a1.125 1.125 0 0 1 .12-1.45l.773-.773a1.125 1.125 0 0 1 1.45-.12l.737.527c.35.25.807.272 1.204.107.397-.165.71-.505.78-.929l.15-.894Z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
            </svg>

        </button>

        <!-- Insights -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Insights'}
            class:dark:bg-gray-200={selectedSection === 'Insights'}
            class:text-gray-500={selectedSection === 'Insights'}
            class:dark:text-gray-500={selectedSection === 'Insights'}
            class:dark:text-gray-400={selectedSection !== 'Insights'}
            class:hover:bg-gray-100={selectedSection !== 'Insights'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Insights'}
            on:click={() => selectSection('Insights')}
            title="Insights"
        >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 0 0 6 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0 1 18 16.5h-2.25m-7.5 0h7.5m-7.5 0-1 3m8.5-3 1 3m0 0 .5 1.5m-.5-1.5h-9.5m0 0-.5 1.5m.75-9 3-3 2.148 2.148A12.061 12.061 0 0 1 16.5 7.605" />
            </svg>

        </button>

        <!-- Terminal -->
        <button 
            class="p-2 rounded-lg transition-colors flex items-center gap-2"
            class:bg-gray-300={selectedSection === 'Terminal'}
            class:dark:bg-gray-200={selectedSection === 'Terminal'}
            class:text-gray-500={selectedSection === 'Terminal'}
            class:dark:text-gray-500={selectedSection === 'Terminal'}
            class:dark:text-gray-400={selectedSection !== 'Terminal'}
            class:hover:bg-gray-100={selectedSection !== 'Terminal'}
            class:dark:hover:bg-gray-800={selectedSection !== 'Terminal'}
            on:click={() => selectSection('Terminal')}
            title="Terminal"
        >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="m6.75 7.5 3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 18V6a2.25 2.25 0 0 0-2.25-2.25H5.25A2.25 2.25 0 0 0 3 6v12a2.25 2.25 0 0 0 2.25 2.25Z" />
            </svg>

        </button>
    </div>
</div>

<!-- Projects Display -->
{#if $viewMode === 'grid'}
  <!-- Grid View -->
  <ProjectGrid {gridCols} />
{:else}
  <!-- List View -->
  <ProjectList />
{/if}

<script lang="ts">
  import { viewMode, showFilterDropdown } from '../../../stores';

  function toggleViewMode(mode: 'grid' | 'list') {
    viewMode.set(mode);
  }
</script>

<!-- Sub Header -->
<div class="relative border-b border-gray-700 px-6 py-3">
  <div class="flex items-center justify-between gap-4">
    <!-- Center: Navigation Tabs (Absolutely Centered) -->
    <div class="absolute left-1/2 transform -translate-x-1/2 flex items-center">
      <div class="flex items-center space-x-3 overflow-hidden">
        <a href="#" class="nav-tab active whitespace-nowrap">Overview</a>
        <a href="#" class="nav-tab whitespace-nowrap">Deployments</a>
        <a href="#" class="nav-tab whitespace-nowrap">Resources</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden sm:block">Domains</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden md:block">Insights</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden lg:block">Console</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden xl:block">Settings</a>
      </div>
    </div>

    

    <!-- Right: Controls (right-aligned, min space) -->
    <div class="flex items-center space-x-2 ml-auto">
      <!-- Grid View Button -->
      <button 
        class="p-2 rounded-lg transition-colors flex items-center"
        class:bg-gray-300={$viewMode === 'grid'}
        class:dark:bg-gray-200={$viewMode === 'grid'}
        class:text-gray-500={$viewMode === 'grid'}
        class:dark:text-gray-500={$viewMode === 'grid'}
        class:dark:text-gray-400={$viewMode === 'grid'}
        class:hover:bg-gray-100={$viewMode !== 'grid'}
        class:dark:hover:bg-gray-800={$viewMode !== 'grid'}
        on:click={() => toggleViewMode('grid')}
        aria-label="Grid view"
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 16.875h3.375m0 0h3.375m-3.375 0V13.5m0 3.375v3.375M6 10.5h2.25a2.25 2.25 0 0 0 2.25-2.25V6a2.25 2.25 0 0 0-2.25-2.25H6A2.25 2.25 0 0 0 3.75 6v2.25A2.25 2.25 0 0 0 6 10.5Zm0 9.75h2.25A2.25 2.25 0 0 0 10.5 18v-2.25a2.25 2.25 0 0 0-2.25-2.25H6a2.25 2.25 0 0 0-2.25 2.25V18A2.25 2.25 0 0 0 6 20.25Zm9.75-9.75H18a2.25 2.25 0 0 0 2.25-2.25V6A2.25 2.25 0 0 0 18 3.75h-2.25A2.25 2.25 0 0 0 13.5 6v2.25a2.25 2.25 0 0 0 2.25 2.25Z" />
        </svg>
      </button>
      
      <!-- List View Button -->
      <button 
        class="p-2 rounded-lg transition-colors flex items-center"
        class:bg-gray-300={$viewMode === 'list'}
        class:dark:bg-gray-200={$viewMode === 'list'}
        class:text-gray-500={$viewMode === 'list'}
        class:dark:text-gray-500={$viewMode === 'list'}
        class:dark:text-gray-400={$viewMode === 'list'}
        class:hover:bg-gray-100={$viewMode !== 'list'}
        class:dark:hover:bg-gray-800={$viewMode !== 'list'}
        on:click={() => toggleViewMode('list')}
        aria-label="List view"
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
        </svg>
      </button>

      <!-- Add New Button -->
      <button class="app-button px-4 py-2 rounded-lg font-medium flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v6m3-3H9m12 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
        Add New 
      </button>
    </div>
  </div>
</div>

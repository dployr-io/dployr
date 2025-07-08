<script lang="ts">
  import { viewMode, showFilterDropdown } from '../../../stores';

  function toggleViewMode(mode: 'grid' | 'list') {
    viewMode.set(mode);
  }

  function toggleFilter() {
    showFilterDropdown.update(show => !show);
  }
</script>

<!-- Sub Header -->
<div class="relative border-b border-gray-700 px-6 py-3">
  <div class="flex items-center justify-between gap-4">
    <!-- Centered search with filter button -->
    <div class="absolute inset-x-0 top-1/2 transform -translate-y-1/2 flex justify-center pointer-events-none">
      <div class="relative flex items-center w-full max-w-lg pointer-events-auto">
        <!-- Search Input -->
        <div class="relative flex-1">
          <input
            type="text"
            placeholder="Search Projects…"
            class="app-input w-full pl-8 pr-4 py-1.5 text-sm rounded-lg outline-none transition-all"
          />
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
            </svg>
          </span>
        </div>
        
        <!-- Filter Button - positioned right of search -->
        <div class="relative ml-2">
          <button 
            class="p-2 rounded-lg transition-colors flex items-center "
            class:bg-gray-300={$showFilterDropdown}
            class:dark:bg-gray-200={$showFilterDropdown}
            class:text-gray-500={$showFilterDropdown}
            class:dark:text-gray-500={$showFilterDropdown}
            class:dark:text-gray-400={$showFilterDropdown}
            on:click|stopPropagation={toggleFilter}
            aria-label="Filter"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75" />
            </svg>
          </button>
          
          <!-- Filter Dropdown -->
          {#if $showFilterDropdown}
            <div class="absolute top-full right-0 mt-2 z-200 card rounded-lg shadow-lg min-w-[200px]">
              <div class="p-4">
                <h3 class="font-semibold mb-3">Filter Projects</h3>
                <div class="space-y-2">
                  <label class="flex items-center">
                    <input type="checkbox" class="mr-2" /> Active Projects
                  </label>
                  <label class="flex items-center">
                    <input type="checkbox" class="mr-2" /> Failed Builds
                  </label>
                  <label class="flex items-center">
                    <input type="checkbox" class="mr-2" /> Recent Updates
                  </label>
                </div>
              </div>
            </div>
          {/if}
        </div>
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

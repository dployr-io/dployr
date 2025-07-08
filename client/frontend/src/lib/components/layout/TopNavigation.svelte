<script lang="ts">
  import { 
    projects, 
    selectedProject, 
    accounts, 
    selectedAccount, 
    showProjectDropdown, 
    showAccountDropdown,
    showFilterDropdown
} from '../../../stores';
  import ThemeToggle from '../ui/ThemeToggle.svelte';

  export let icon: string;
  export let iconSecondary: string;
  export let isDarkMode: boolean;

  function selectProject(project: any) {
    selectedProject.set(project);
    showProjectDropdown.set(false);
  }

  function selectAccount(account: any) {
    selectedAccount.set(account);
    showAccountDropdown.set(false);
  }

  function toggleProjectDropdown() {
    showProjectDropdown.update(show => !show);
    showAccountDropdown.set(false);
  }

  function toggleAccountDropdown(e: { stopPropagation: () => void; }) {
    e.stopPropagation();
    showAccountDropdown.update(show => !show);
    showProjectDropdown.set(false);
  }

  function toggleFilter() {
    showFilterDropdown.update(show => !show);
  }
</script>

<nav class="border-b border-gray-700 flex-shrink-0">
  <div class="relative flex items-center justify-between px-6 py-3">
    <!-- Left: Logo and Project/Account Dropdown -->
    <div class="flex items-center space-x-3 relative flex-shrink-0 min-w-0">
      <img 
        alt="App logo" 
        src="{isDarkMode ? iconSecondary : icon}"
        class="w-8 h-8 rounded-xl flex-shrink-0"
      />

      <!-- Project Dropdown Trigger -->
      <button 
        class="text-white font-semibold flex items-center space-x-1 focus:outline-none min-w-0"
        on:click|stopPropagation={toggleProjectDropdown}
      >
        <span class="text-gray-600 dark:text-gray-200 truncate min-w-0">John's workspace</span>
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 text-gray-600 dark:text-gray-200 flex-shrink-0">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 15 12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9" />
        </svg>
      </button>

      <!-- Project Dropdown -->
      {#if $showProjectDropdown}
        <div class="absolute left-0 top-0 mt-9 z-50 bg-white rounded-lg shadow-lg min-w-[220px]">
          <div class="p-2">
            {#each $projects as project}
              <button 
                class="group flex items-center w-full px-3 py-2 rounded app-button-ghost text-left hover:text-white active:text-white"
                on:click={() => selectProject(project)}
              >
                <img src={project.icon} alt="icon" class="w-7 h-7 rounded mr-2"/>
                <div>
                  <!-- use group-hover and group-active to override the base gray -->
                  <div class="text-gray-600 group-hover:text-white group-active:text-white font-medium text-sm">
                    {project.name}
                  </div>
                </div>
              </button>
            {/each}
            <div class="border-t border-gray-20 my-2"></div>
              <div class="flex justify-center">
                <button
                  class="w-full max-w-[180px] px-3 py-1 text-sm text-left font-medium
                        text-gray-600"
                >
                  + New Project
                </button>
              </div>
          </div>
          <div class="border-t border-gray-200"></div>
          <div class="p-2">
            <button 
              class="group flex items-center w-full px-3 py-1 rounded app-button-ghost text-left hover:text-white active:text-white"
              on:click|stopPropagation={toggleAccountDropdown}
            >
              <span class="font-semibold text-sm text-gray-600 group-hover:text-white group-active:text-white">
                Switch Account
              </span>
              <svg 
                class="w-4 h-4 ml-auto text-gray-600 group-hover:text-white group-active:text-white" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>
            {#if $showAccountDropdown}
              <div class="absolute left-full top-0 ml-2 z-50 bg-white dark:bg-gray-800 border border-gray-200 rounded-lg shadow-lg min-w-[180px]">
                <div class="p-2">
                  <div class="text-xs text-gray-500 px-2 py-1">Accounts</div>
                  {#each $accounts as account}
                    <button 
                      class="flex items-center w-full px-3 py-2 rounded app-button-ghost text-left"
                      on:click={() => selectAccount(account)}
                    >
                      <div class="w-6 h-6 rounded-full bg-gray-300 dark:bg-gray-600 mr-2"></div>
                      <span>{account.name}</span>
                    </button>
                  {/each}
                  <div class="border-t border-gray-200 dark:border-gray-700 my-2"></div>
                  <button class="w-full px-3 py-2 text-left text-blue-600 hover:underline">+ New Account</button>
                </div>
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </div>

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

    <!-- Right: User Controls -->
    <div class="flex items-center space-x-3 flex-shrink-0">
      <img src="https://picsum.photos/200/20" alt="Profile" class="h-7 w-7 rounded-full" />
      <div>
        |
      </div>
      <button 
        class="flex w-8 h-8 rounded-lg dark:bg-white/10 bg-white/40 hover:bg-white/20 items-center justify-center border border-white/20"
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0" />
        </svg>
      </button>
      <ThemeToggle />
    </div>
  </div>
</nav>

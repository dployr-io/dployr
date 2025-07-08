<script lang="ts">
  import { 
    projects, 
    selectedProject, 
    accounts, 
    selectedAccount, 
    showProjectDropdown, 
    showAccountDropdown 
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
        <div class="absolute left-0 top-0 mt-9 z-50 card rounded-lg shadow-lg min-w-[220px]">
          <div class="p-2">
            <div class="text-sm font-semibol px-2 py-1 w-fit text-left">{'Projects'}</div>
            {#each $projects as project}
              <button 
                class="flex items-center w-full px-3 py-2 rounded app-button-ghost text-left"
                on:click={() => selectProject(project)}
              >
                <div class="w-7 h-7 rounded bg-gray-200 dark:bg-gray-700 flex items-center justify-center mr-2 font-bold">{project.icon}</div>
                <div>
                  <div class="text-gray-600 dark:text-gray-200 font-medium">{project.name}</div>
                </div>
              </button>
            {/each}
            <div class="border-t border-gray-200 dark:border-gray-700 my-2"></div>
              <div class="flex justify-center">
              <button class="w-full max-w-[180px] px-3 py-2 text-left text-blue-600 hover:underline">+ New Project</button>
              </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-700"></div>
          <div class="p-2">
            <button 
              class="flex items-center w-full px-3 py-2 rounded app-button-ghost text-left"
              on:click|stopPropagation={toggleAccountDropdown}
            >
              <span class="font-semibold text-gray-600 dark:text-gray-200">Switch Account</span>
              <svg class="w-4 h-4 ml-auto text-gray-600 dark:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>
            {#if $showAccountDropdown}
              <div class="absolute left-full top-0 ml-2 z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg min-w-[180px]">
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

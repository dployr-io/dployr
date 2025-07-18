<script lang="ts">
  import TopNavigation from './TopNavigation.svelte';
  import SubHeader from './SubHeader.svelte';
  import MainContent from './MainContent.svelte';
  import { 
    sidebarWidth, 
    isResizing, 
    showFilterDropdown, 
    showProjectDropdown, 
    showAccountDropdown, 
    showProfileDropdown, 
    showNewProjectPopup, 
    host,

    token,

    name,

    isLoading




  } from '../../../stores';
  import { SIDEBAR_WIDTH_DOCKED } from '../../../../src/constants';
  import ProjectGrid from '../project/ProjectGrid.svelte';
  import Modal from '../ui/Modal.svelte';
  import { projectService } from '../../../../src/lib/services/api';
  import { setTimeout } from 'timers/promises';
  
  export let icon: string;
  export let iconSecondary: string;
  export let isDarkMode: boolean;

  export let projectName: string = "";
  export let gitRepo: string = "";
  export let error: Error;

  $: isDisabled = projectName.length > 3 && gitRepo.length > 3;

  function startResize(e: MouseEvent) {
    isResizing.set(true);
    e.preventDefault();
    
    function handleMouseMove(e: MouseEvent) {
      if (!$isResizing) return;
      const newWidth = Math.max(600, Math.min(800, e.clientX));
      sidebarWidth.set(newWidth);
    }
    
    function handleMouseUp() {
      isResizing.set(false);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    }
    
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }

  // Hide dropdowns on outside click
  if (typeof window !== 'undefined') {
    window.addEventListener('click', () => {
      showProjectDropdown.set(false);
      showAccountDropdown.set(false);
      showFilterDropdown.set(false);
      showProfileDropdown.set(false);
    });
  }
  
  async function handleCreateProject() {
   isLoading.set(true);
    try {
      projectService.createProject($host, $token, {
        "git_repo": gitRepo,
        "name": projectName
      })
      gitRepo = "";
      projectName = "";
    } catch (e) {
      error = e as Error;
    } finally {
       isLoading.set(false);
    }
  }
</script>

<div class="w-full h-screen flex flex-col">
  {#if $showNewProjectPopup} 
    <Modal bind:show={$showNewProjectPopup} title="New Project">
      <label for="project-name" class="block text-sm font-semibold text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Name</label>
      <input 
        id="project-name"
        name="project-name"
        type="text" 
        placeholder="Enter project name" 
        bind:value={projectName} 
        class="app-input font-medium w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
      />

      <div class="h-4" />

      <label for="git-repo" class="block text-sm font-semibold text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Remote repository</label>
      <input 
        id="git-repo"
        name="git-repo"
        type="text" 
        placeholder="Enter link to git repository" 
        bind:value={gitRepo} 
        class="app-input font-medium w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
      />

      <div class="h-4" />
      {#if error}
        <div class="flex pl-4">
          <p class="font-semibold text-sm text-left text-red-400">
            {error.message}
          </p>
        </div>
      {/if}

      <div slot="footer" class="flex gap-3">
        <button class="app-button-outlined px-4 py-2 rounded-lg" on:click={() => $showNewProjectPopup = false}>Cancel</button>
        <button  
          class="
            px-4 py-2 rounded-lg font-semibold text-sm
            bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]
            disabled:bg-[#CFDBD5] disabled:text-gray-500 disabled:cursor-not-allowed
          "
          disabled={!isDisabled}
          on:click={handleCreateProject}
        >
          {#if $isLoading}
            <div class="flex items-center gap-2">
              <div class="w-5 h-5 border-2 border-gray-500 border-t-transparent rounded-full animate-spin mx-auto" />
              Loading
            </div>
          {:else}
            Create Project
          {/if}
        </button>
      </div>
    </Modal>
  {/if}

  <TopNavigation {icon} {iconSecondary} {isDarkMode} />
  <SubHeader />
  
  <!-- Main Content -->
  <div class="flex flex-1 min-h-0">
    <!-- Left Sidebar -->
    <div class="sidebar border-r border-gray-700 flex-shrink-0" style="width: {$sidebarWidth}px;"
      class:p-6={$sidebarWidth !== SIDEBAR_WIDTH_DOCKED}
      class:pt-6={$sidebarWidth === SIDEBAR_WIDTH_DOCKED}
      class:px-2={$sidebarWidth === SIDEBAR_WIDTH_DOCKED}
    >
      <ProjectGrid />
    </div>

    <!-- Resize Handle -->
    <button 
      class="w-1 cursor-col-resize transition-colors flex-shrink-0 relative flex items-center justify-center border-0 bg-transparent"
      on:mousedown={startResize}
      class:bg-blue-500={$isResizing}
      aria-label="Resize sidebar"
    >
      <!-- Capsule Handle Visual Indicator -->
      <div 
        class="absolute left-[-2px] w-1 h-8 bg-gray-400 dark:bg-gray-500 rounded-full hover:bg-gray-500 dark:hover:bg-gray-400 transition-colors"
        class:bg-blue-500={$isResizing}
        class:hover:bg-blue-400={$isResizing}
        class:w-2={$isResizing}
      >
      </div>
    </button>

    <!-- Main Content Area -->
    <div class="flex-1 p-6 min-h-0 overflow-auto">
      <MainContent />
    </div>
  </div>
</div>

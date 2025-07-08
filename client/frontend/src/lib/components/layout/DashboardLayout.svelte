<script lang="ts">
  import TopNavigation from './TopNavigation.svelte';
  import SubHeader from './SubHeader.svelte';
  import Sidebar from './Sidebar.svelte';
  import MainContent from './MainContent.svelte';
  import { sidebarWidth, isResizing, showFilterDropdown, showProjectDropdown, showAccountDropdown } from '../../../stores';
  
  export let icon: string;
  export let iconSecondary: string;
  export let isDarkMode: boolean;

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
    });
  }
</script>

<div class="w-full h-screen flex flex-col">
  <TopNavigation {icon} {iconSecondary} {isDarkMode} />
  <SubHeader />
  
  <!-- Main Content -->
  <div class="flex flex-1 min-h-0">
    <!-- Left Sidebar -->
    <div class="sidebar border-r border-gray-700 p-6 flex-shrink-0" style="width: {$sidebarWidth}px;">
      <Sidebar />
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
      <MainContent sidebarWidth={$sidebarWidth} />
    </div>
  </div>
</div>

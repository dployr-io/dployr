<script lang="ts">
  import { viewMode } from '../../../stores';
  import ProjectGrid from '../project/ProjectGrid.svelte';
  
  export let sidebarWidth: number;

  // Calculate grid columns based on available width
  $: mainContentWidth = typeof window !== 'undefined' ? window.innerWidth - sidebarWidth - 48 : 800; // 48px for padding
  $: gridCols = Math.max(1, Math.floor(mainContentWidth / 350)); // 350px min card width
  $: maxGridCols = 3;
</script>

<!-- Projects Display -->
{#if $viewMode === 'grid'}
  <!-- Grid View -->
  <ProjectGrid gridCols={gridCols} maxGridCols={maxGridCols} />
{:else}
  <!-- List View -->
  <ProjectGrid gridCols={gridCols} maxGridCols={1}  />
{/if}

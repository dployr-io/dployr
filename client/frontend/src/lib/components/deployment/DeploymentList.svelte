<script lang="ts">
  import { deployments, appState } from '../../../stores';
  import DeploymentCard from './DeploymentCard.svelte';

  function selectDeployment(deployment: any) {
    appState.update(state => ({ ...state, selectedDeploymentId: deployment.id }));
  }
</script>

<div class="flex flex-col gap-6 h-full max-w-5xl">
    <!-- Left Filters Sidebar -->
    <div class="flex gap-2">
        <!-- Branch Filter -->
        <div class="min-w-0 flex-1">
            <select id="branch-filter" class="app-input w-full px-3 py-1 rounded text-sm truncate">
                <option class="text-gray-500 font-medium">All Branches</option>
                <option class="text-gray-500 font-medium">main</option>
                <option class="text-gray-500 font-medium">develop</option>
            </select>
        </div>
        
        <!-- Date Range -->
        <div class="min-w-0 flex-1">
            <button id="date-range" class="app-input w-full px-3 py-1 rounded text-left text-sm truncate">Date Range</button>
        </div>
        
        <!-- Environment -->
        <div class="min-w-0 flex-1">
            <select id="environment-filter" class="app-input w-full px-3 py-1 rounded text-sm truncate">
                <option class="text-gray-500 font-medium">All Environments</option>
                <option class="text-gray-500 font-medium">Production</option>
                <option class="text-gray-500 font-medium">Staging</option>
            </select>
        </div>
        
        <!-- Status -->
        <div class="min-w-0 flex-1">
            <select id="status-filter" class="app-input w-full px-3 py-1 rounded text-sm truncate">
                <option class="text-gray-500 font-medium">All Status</option>
                <option class="text-gray-500 font-medium">Ready</option>
                <option class="text-gray-500 font-medium">Error</option>
                <option class="text-gray-500 font-medium">Building</option>
            </select>
        </div>
    </div>
    
    <!-- Main Content -->
    <div class="flex-1 space-y-4">
        <div class="flex items-center justify-between">
            <div class="text-sm text-gray-500">{$deployments.length}/{$deployments.length} deployments</div>
            <button class="app-button px-4 py-2 rounded-lg">Deploy Now</button>
        </div>
        
        <!-- Deployments List -->
        <div class="space-y-3">
          {#each $deployments as deployment}
            <DeploymentCard 
              {deployment} 
              isSelected={$appState.selectedDeploymentId === deployment.id}
              on:select={() => selectDeployment(deployment)}
            />
          {/each}
        </div>
    </div>
</div>

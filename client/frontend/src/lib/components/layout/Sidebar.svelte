<script lang="ts">
  import { appState, deployments } from '../../../stores';
  import DeploymentList from '../deployment/DeploymentList.svelte';
  import LogsView from '../deployment/LogsView.svelte';
  import DomainsView from '../deployment/DomainsView.svelte';
  import ConsoleView from '../deployment/ConsoleView.svelte';
</script>

<!-- Section Content -->
<div class="h-full flex flex-col">
  {#if $appState.selectedSection === 'Deployments'}
    <DeploymentList />
  {:else if $appState.selectedSection === 'Logs'}
    <LogsView />
  {:else if $appState.selectedSection === 'Resources'}
    <!-- Resources content -->
    <div class="text-gray-500">Resources section coming soon...</div>
  {:else if $appState.selectedSection === 'Domains'}
    <!-- Domains content -->
    <DomainsView />
  {:else if $appState.selectedSection === 'Settings'}
    <!-- Settings content -->
    <div class="text-gray-500">Settings section coming soon...</div>
  {:else if $appState.selectedSection === 'Insights'}
    <!-- Insights content -->
    <div class="text-gray-500">Insights section coming soon...</div>
  {:else if $appState.selectedSection === 'Console'}
    <!-- Terminal content -->
    <ConsoleView />
  {/if}
</div>

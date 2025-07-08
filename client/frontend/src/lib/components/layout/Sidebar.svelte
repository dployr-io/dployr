<script lang="ts">
  import { appState, deployments } from '../../../stores';
  import DeploymentList from '../deployment/DeploymentList.svelte';
  import LogsView from '../deployment/LogsView.svelte';

  let selectedSection = 'Deployments';

  function selectSection(section: string) {
    selectedSection = section;
    appState.update(state => ({ ...state, selectedSection: section }));
  }

  $: $appState.selectedSection = selectedSection;
</script>

<!-- Section Content -->
<div class="min-h-0 flex-1">
  {#if selectedSection === 'Deployments'}
    <DeploymentList />
  {:else if selectedSection === 'Logs'}
    <LogsView />
  {:else if selectedSection === 'Resources'}
    <!-- Resources content -->
    <div class="text-gray-500">Resources section coming soon...</div>
  {:else if selectedSection === 'Domains'}
    <!-- Domains content -->
    <div class="text-gray-500">Domains section coming soon...</div>
  {:else if selectedSection === 'Settings'}
    <!-- Settings content -->
    <div class="text-gray-500">Settings section coming soon...</div>
  {:else if selectedSection === 'Insights'}
    <!-- Insights content -->
    <div class="text-gray-500">Insights section coming soon...</div>
  {:else if selectedSection === 'Terminal'}
    <!-- Terminal content -->
    <div class="text-gray-500">Terminal section coming soon...</div>
  {/if}
</div>

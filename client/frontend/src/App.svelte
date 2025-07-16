<script lang="ts">
  import logo from './assets/images/logo.png';
  import logoSecondary from './assets/images/logo-secondary.png';
  import icon from './assets/images/icon.svg';
  import iconSecondary from './assets/images/icon-secondary.svg';
  import { onMount, onDestroy } from 'svelte';
  import ToastContainer from '../src/lib/components/ui/ToastContainer.svelte';
  
  // Components
  import { OnboardingFlow, DashboardLayout } from './lib/components';
  
  // Store
  import { 
    currentUser, 
    isDarkMode, 
    deployments, 
    projects, 
    selectedProject, 
    logs,
    domains,
    wsconsole,
    token,
  } from './stores';
  
  // Service
  import { authService, dataService } from './lib/services/api';
  import { types } from '../wailsjs/go/models';
  import { addToast } from './stores/toastStore';

  let terminalComponent;

  async function handleSignOut() {
    try {
      await authService.signOut();
      currentUser.set(null);
      // Delete locally saved data and configs
      // Optionally navigate to sign-in page
    } catch (error) {
      console.error('Sign out error:', error);
    }
  }

  async function loadData() {
    try {
      const [deploymentData, projectData, logData, domainsData, consoleData] = await Promise.all([
        dataService.getDeployments(),
        dataService.getProjects(),
        dataService.getLogs(),
        dataService.getDomains(),
        dataService.newConsole(),
      ]);

      deployments.set(deploymentData);
      projects.set(projectData);
      logs.set(logData);
      domains.set(domainsData);
      wsconsole.set(consoleData);

      const [tokenData, userData] = await Promise.all([
        authService.getToken(),
        authService.getCurrentUser(),
      ]);
      
      token.set(tokenData);
      currentUser.set(userData);
    } catch (error) {
      console.error('Failed to load data:', error);
    }
  }

  onMount(() => {
    loadData();
  });

  // Auto-select first project when projects load
  $: if ($projects.length > 0 && !$selectedProject) {
    selectedProject.set($projects[0]);
  }
</script>

<ToastContainer />

<main class="w-full flex items-center justify-center min-h-screen">
  {#if !$currentUser}
    <OnboardingFlow {logo} {logoSecondary} isDarkMode={$isDarkMode} />
  {:else}
    <DashboardLayout {icon} {iconSecondary} isDarkMode={$isDarkMode} />
  {/if}
</main>
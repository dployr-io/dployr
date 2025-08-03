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
    host,
  } from './stores';
  
  // Service
  import { authService, consoleService, dataService } from './lib/services/api';

  async function loadData() {
    if (!$currentUser) return;

    try {
      const [tokenData, userData, hostData] = await Promise.all([
        authService.getToken(),
        authService.getCurrentUser(),
        authService.getHost(),
      ]);
      
      token.set(tokenData as string);
      currentUser.set(userData);
      host.set(hostData as string);

      const [deploymentData, projectData, logData, domainsData] = await Promise.all([
        dataService.getDeployments(),
        dataService.getProjects(hostData || '', tokenData || ''),
        dataService.getLogs(),
        dataService.getDomains(),
      ]);

      deployments.set(deploymentData);
      projects.set(projectData);
      logs.set(logData);
      domains.set(domainsData);

      const consoleData = await consoleService.newConsole();
      wsconsole.set(consoleData);

      // Auto-select first project when projects load
      if (projectData.length > 0 && !$selectedProject) {
        selectedProject.set(projectData[0]);
      }
    } catch (error) {
      console.error('Failed to load data:', error);
    }
  }

  onMount(() => {
    loadData();
  
    let interval = setInterval(loadData, 30000);
    
    const handleVisibilityChange = () => {
      if (document.hidden) {
        clearInterval(interval);
      } else {
        loadData();
        interval = setInterval(loadData, 30000);
      }
    };
    
    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    return () => {
      clearInterval(interval);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  });
</script>

<ToastContainer />

<main class="w-full flex items-center justify-center min-h-screen">
  {#if !$currentUser}
    <OnboardingFlow {logo} {logoSecondary} isDarkMode={$isDarkMode} />
  {:else}
    <DashboardLayout {icon} {iconSecondary} isDarkMode={$isDarkMode} />
  {/if}
</main>
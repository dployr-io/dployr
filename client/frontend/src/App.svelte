<script lang="ts">
  import logo from './assets/images/logo.png';
  import logoSecondary from './assets/images/logo-secondary.png';
  import icon from './assets/images/icon.svg';
  import iconSecondary from './assets/images/icon-secondary.svg';
  import { onMount } from 'svelte';
  
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

    domains


  } from './stores';
  
  // Service
  import { authService, dataService } from './lib/services/api';
  import { main } from '../wailsjs/go/models';

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

  async function checkAuthStatus() {
    // const user = await authService.getCurrentUser();

    // DEV FEATURE: to be removed
    const user = new main.User({
      id: "78329839993909317",
      email: "john.doe@example.com",
      name: "John Doe",
      avatar: "https://picsum.photos/200/200"
    })

    currentUser.set(user);
  }

  async function loadData() {
    try {
      const [deploymentData, projectData, logData, domainsData] = await Promise.all([
        dataService.getDeployments(),
        dataService.getProjects(),
        dataService.getLogs(),
        dataService.getDomains(),
      ]);
      deployments.set(deploymentData);
      projects.set(projectData);
      logs.set(logData);
      domains.set(domainsData);
    } catch (error) {
      console.error('Failed to load data:', error);
    }
  }

  onMount(() => {
    checkAuthStatus();
    loadData();
  });

  // Auto-select first project when projects load
  $: if ($projects.length > 0 && !$selectedProject) {
    selectedProject.set($projects[0]);
  }
</script>

<main class="w-full flex items-center justify-center min-h-screen">
  {#if !$currentUser}
    <OnboardingFlow {logo} {logoSecondary} isDarkMode={$isDarkMode} />
  {:else}
    <DashboardLayout {icon} {iconSecondary} isDarkMode={$isDarkMode} />
  {/if}
</main>
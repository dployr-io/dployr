<script lang="ts">
  import logo from './assets/images/logo.png';
  import logoSecondary from './assets/images/logo-secondary.png';
  import icon from './assets/images/icon.svg';
  import iconSecondary from './assets/images/icon-secondary.svg';
  import { onMount } from 'svelte';
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
    isAuthInitialized,
    isAuthenticating,
  } from './stores';
  
  // Service
  import { authService, consoleService, dataService } from './lib/services/api';

  async function initializeAuth() {
    try {
      isAuthenticating.set(true);
      
      // Initialize authentication state from localStorage
      const authResult = await authService.initializeAuthState();
      
      // Update stores with restored authentication data
      if (authResult.isAuthenticated) {
        currentUser.set(authResult.user);
        token.set(authResult.token || '');
        host.set(authResult.host || '');
      } else {
        // Clear stores if not authenticated
        currentUser.set(null);
        token.set('');
        host.set('');
      }
      
      // Mark authentication as initialized
      isAuthInitialized.set(true);
      
      return authResult.isAuthenticated;
    } catch (error) {
      console.error('Failed to initialize authentication:', error);
      
      // Clear stores on error
      currentUser.set(null);
      token.set('');
      host.set('');
      isAuthInitialized.set(true);
      
      return false;
    } finally {
      isAuthenticating.set(false);
    }
  }

  async function loadData() {
    // Ensure we have authentication data before proceeding
    if (!$currentUser || !$token || !$host) {
      console.warn('Cannot load data: missing authentication data');
      return;
    }

    try {
      // Use the already-initialized authentication data from stores
      const currentHost = $host;
      const currentToken = $token;

      const [deploymentData, projectData, logData, domainsData] = await Promise.all([
        dataService.getDeployments(currentHost, currentToken),
        dataService.getProjects(currentHost, currentToken),
        dataService.getLogs(currentHost, currentToken),
        dataService.getDomains(currentHost, currentToken),
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
    let interval: any;
    
    const initialize = async () => {
      // Initialize authentication state first
      const isAuthenticated = await initializeAuth();
      
      // Only load data if authenticated
      if (isAuthenticated) {
        await loadData();
      }
    
      interval = setInterval(() => {
        if ($currentUser) {
          loadData();
        }
      }, 30000);
    };
    
    const handleVisibilityChange = () => {
      if (document.hidden) {
        clearInterval(interval);
      } else {
        if ($currentUser) {
          loadData();
          interval = setInterval(() => {
            if ($currentUser) {
              loadData();
            }
          }, 30000);
        }
      }
    };
    
    // Start initialization
    initialize();
    
    document.addEventListener('visibilitychange', handleVisibilityChange);
    
    return () => {
      clearInterval(interval);
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  });
</script>

<ToastContainer />

<main class="w-full flex items-center justify-center min-h-screen">
  {#if !$isAuthInitialized || $isAuthenticating}
    <!-- Loading state during authentication initialization -->
    <div class="flex flex-col items-center justify-center space-y-4">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="text-sm text-gray-600 dark:text-gray-400">Initializing...</p>
    </div>
  {:else if !$currentUser}
    <OnboardingFlow {logo} {logoSecondary} isDarkMode={$isDarkMode} />
  {:else}
    <DashboardLayout {icon} {iconSecondary} isDarkMode={$isDarkMode} />
  {/if}
</main>
<script lang="ts">
  import { saveToLocalStorage } from '../../../../src/utils/localStorage';
  import { addToast } from '../../../../src/stores/toastStore';
  import { 
    currentPage, 
    isTransitioning, 
    selectedOptions, 
    discoveryOptions, 
    discoveryOther, 
    appStage, 
    currentUser,
    host,
    email,
    password,
    name,
    privateKey,
    isLoading,
    otp,
} from '../../../stores';
  import { getErrorMessage } from '../../../utils';
  import { authService } from '../../services/api';
  import OnboardingPage from './OnboardingPage.svelte';
  
  export let logo: string;
  export let logoSecondary: string;
  export let isDarkMode: boolean;
  const pages = [
    {
      title: "How do you intend to use dployr?",
      options: [
        "Deploy applications to any infrastructure",
        "Design new cloud architecture",
        "Learn deployment and orchestration", 
        "Explore platform capabilities"
      ]
    },
    {
      title: "How did you find dployr?",
      options: [
        "GitHub",
        "Twitter / X",
        "Hacker News",
        "Reddit",
        "Product Hunt",
        "Dev.to / Daily.dev / Medium",
        "Search engine (Google, etc.)",
        "Word of mouth / Friend",
        "Conference or meetup",
        "Other"
      ]
    },
    {
      title: "What is the stage of your app?",
      options: [
        "New project idea",
        "Existing project"
      ]
    },
    {
      title: "Sign in your account",
    },
    {
      title: "Verify your email",
    }
  ];

  function toggleOption(option: string) {
    selectedOptions.update(opts => {
      if (opts.includes(option)) {
        return opts.filter(opt => opt !== option);
      } else {
        return [...opts, option];
      }
    });
  }

  function toggleDiscovery(option: string) {
    discoveryOptions.update(opts => {
      if (opts.includes(option)) {
        if (option === 'Other') discoveryOther.set('');
        return opts.filter(opt => opt !== option);
      } else {
        return [option];
      }
    });
  }

  function selectAppStage(option: string) {
    appStage.set(option);
  }

  async function handleSignIn(host: string, email: string, name: string, password: string, privateKey: string) {
   try {
    $isLoading = true;
    const res = await authService.signIn(host, email, name, password, privateKey);
    addToast(res.message, 'success');
    nextPage();
   } catch (error) {
    addToast(getErrorMessage(error), 'error');
   } finally {
    $isLoading = false;
   }
  }

  async function handleMagicCode(host: string, email: string, code: string) {
    try{
      $isLoading = true;
      const res = await authService.verifyMagicCode(host, email, code);      
      $currentUser = res.user;

      saveToLocalStorage("user", res.user);
      saveToLocalStorage("token", res.token);
    } catch (error) {
      addToast(getErrorMessage(error), 'error');
    } finally {
      $isLoading = false;
    }
  }

  $: canProceed = (() => {
    if ($currentPage === 0) return $selectedOptions.length > 0;
    if ($currentPage === 1) return $discoveryOptions.length > 0 && (!$discoveryOptions.includes('Other') || $discoveryOther.trim().length > 0);
    if ($currentPage === 2) return !!$appStage;
    if ($currentPage === 3) return !!$host && !!$email && !!$name;
    if ($currentPage === 4) return $otp.length === 6;
    return false;
  })();

  function canGoNext() {
    return canProceed;
  }

  function nextPage() {
    if (!canGoNext() || $isTransitioning) return;
    isTransitioning.set(true);
    setTimeout(() => {
      currentPage.update(p => Math.min(p + 1, pages.length - 1));
      isTransitioning.set(false);
    }, 150);
  }

  function previousPage() {
    if ($isTransitioning) return;
    isTransitioning.set(true);
    setTimeout(() => {
      currentPage.update(p => Math.max(p - 1, 0));
      isTransitioning.set(false);
    }, 150);
  }
</script>

<!-- Progress bar -->
<div class="fixed top-0 left-0 right-0 h-1 z-50 bg-[#CFDBD5]">
  <div 
    class="h-full transition-all duration-300 ease-in-out bg-[#195B5E]"
    style="width: {(($currentPage + 1) / pages.length) * 100}%"
  ></div>
</div>

<main class="w-full flex items-center justify-center min-h-screen">
  <div class="rounded-xl p-12 transition-all duration-300 ease-in-out {$isTransitioning ? 'opacity-70 translate-x-5' : 'opacity-100 translate-x-0'}">
    <div class="flex flex-col items-center gap-6 mb-8">
      <img 
        alt="App logo" 
        src="{isDarkMode ? logoSecondary : logo}"
        class="w-32 rounded-xl flex-shrink-0"
      >
      <div class="text-xl font-semibold text-center flex-1 leading-relaxed text-gray-700 dark:text-gray-100">
        {#if !$currentUser}
          {pages[$currentPage].title}
        {:else}
          <p>Welcome back, {$currentUser.name}!</p>
        {/if}
      </div>

      {#if !$currentUser && $currentPage === 3}
        <div class="flex flex-col items-start">
          <p>To get started ensure you have setup your server.</p>  
          <p><a href="" class="underline font-semibold">Click</a> to learn how.</p>  
        </div>
      {/if}
    </div>
    
    <OnboardingPage 
      {pages}
      {toggleOption}
      {toggleDiscovery}
      {selectAppStage}
      handleSignIn={async () =>
        handleSignIn(
          $host,     
          $email,
          $name,
          $password,
          $privateKey
      )}
      handleMagicCode={async () =>
        handleMagicCode(
          $host, $email, $otp,
      )}
      {nextPage}
      {previousPage}
      {canProceed}
      currentPage={$currentPage}
      selectedOptions={$selectedOptions}
      discoveryOptions={$discoveryOptions}
      appStage={$appStage}
    />
  </div>
</main>

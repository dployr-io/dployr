<script lang="ts">
  import logo from './assets/images/logo.png'
  import logoSecondary from './assets/images/logo-secondary.png'
  import {SignIn, GetCurrentUser, SignOut} from '../wailsjs/go/main/App.js'
  import { onMount } from 'svelte'
  import { main } from '../wailsjs/go/models';
  
  let currentTheme: 'light' | 'dark' | 'system' = 'system'
  let isDarkMode = false
  let currentUser: main.User | null = null
  let isAuthenticating = false
  let selectedOptions: string[] = []
  let discoveryOptions: string[] = []
  let discoveryOther: string = ''
  let appStage: string = ''
  let signInProvider: string = ''
  let currentPage = 0
  let isTransitioning = false

  const pages = [
    {
      title: "How do you intend to use dployr?",
      options: [
        "Deploy applications to any infrastructure",
        "Design custom cloud architectures",
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
      options: [
        "GitHub",
        "GitLab",
        "Bitbucket",
        "Unity",
      ]
    }
  ];

  function toggleOption(option: string) {
    if (selectedOptions.includes(option)) {
      selectedOptions = selectedOptions.filter(opt => opt !== option)
    } else {
      selectedOptions = [...selectedOptions, option]
    }
  }

  function toggleDiscovery(option: string) {
    if (discoveryOptions.includes(option)) {
      discoveryOptions = discoveryOptions.filter(opt => opt !== option)
      if (option === 'Other') discoveryOther = ''
    } else {
      discoveryOptions = [option]
    }
  }

  function selectAppStage(option: string) {
    appStage = option
  }

  async function handleSignIn(provider: string) {
    if (isAuthenticating) return
    isAuthenticating = true

    const result = await SignIn(provider.toLowerCase())
    if (!result.Success) {
        console.error('Failed to open browser')
        return
    }
    
    const pollInterval = setInterval(async () => {
        const user = await GetCurrentUser()
        if (user) {
            clearInterval(pollInterval)
            currentUser = user
            isAuthenticating = false
        }
    }, 2000) // Check every 2 seconds

    // Timeout after 3 minutes
    setTimeout(() => {
        clearInterval(pollInterval)
        console.log('Authentication timeout')
    }, 180000)
  }

  async function handleSignOut() {
    try {
      await SignOut()
      currentUser = null
      // Optionally navigate to sign-in page
    } catch (error) {
      console.error('Sign out error:', error)
    }
  }

  async function checkAuthStatus() {
    currentUser = await GetCurrentUser()
  }

  $: canProceed = (() => {
    if (currentPage === 0) return selectedOptions.length > 0
    if (currentPage === 1) return discoveryOptions.length > 0 && (!discoveryOptions.includes('Other') || discoveryOther.trim().length > 0)
    if (currentPage === 2) return !!appStage
    if (currentPage === 3) return !!signInProvider
    return false
  })()

  function canGoNext() {
    return canProceed
  }

  function nextPage() {
    if (!canGoNext() || isTransitioning) return
    isTransitioning = true
    setTimeout(() => {
      currentPage = Math.min(currentPage + 1, pages.length - 1)
      isTransitioning = false
    }, 150)
  }

  function previousPage() {
    if (isTransitioning) return
    isTransitioning = true
    setTimeout(() => {
      currentPage = Math.max(currentPage - 1, 0)
      isTransitioning = false
    }, 150)
  }

  function getSystemTheme(): boolean {
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  }

  function applyTheme(theme: 'light' | 'dark' | 'system') {
    const isDark = theme === 'dark' || (theme === 'system' && getSystemTheme())
    isDarkMode = isDark
    if (isDark) {
      document.documentElement.setAttribute('data-theme', 'dark')
    } else {
      document.documentElement.removeAttribute('data-theme')
    }
  }

  function toggleTheme() {
    const themes: ('light' | 'dark' | 'system')[] = ['light', 'dark', 'system']
    const currentIndex = themes.indexOf(currentTheme)
    const nextIndex = (currentIndex + 1) % themes.length
    currentTheme = themes[nextIndex]
    applyTheme(currentTheme)
  }

  function getThemeIcon() {
    if (currentTheme === 'system') {
      return `<svg viewBox="0 0 24 24" class="w-5 h-5">
        <path d="M12 2A10 10 0 0 0 2 12a10 10 0 0 0 10 10 10 10 0 0 0 10-10A10 10 0 0 0 12 2zm0 18a8 8 0 0 1-8-8 8 8 0 0 1 8-8 8 8 0 0 1 8 8 8 8 0 0 1-8 8zm0-16a8 8 0 0 0-8 8h8V4z"/>
      </svg>`
    } else if (currentTheme === 'dark') {
      return `<svg viewBox="0 0 24 24" class="w-5 h-5">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
      </svg>`
    } else {
      return `<svg viewBox="0 0 24 24" class="w-5 h-5">
        <circle cx="12" cy="12" r="5"/>
        <path d="m12 1 0 2m0 18 0 2M4.22 4.22l1.42 1.42m12.72 12.72 1.42 1.42M1 12l2 0m18 0 2 0M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
      </svg>`
    }
  }

  function getProviderIcon(provider: string) {
    const iconClass = "w-5 h-5 mr-3 flex-shrink-0"
    switch (provider) {
      case 'GitHub':
        return `<svg viewBox="0 0 24 24" class="${iconClass}" fill="currentColor">
          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
        </svg>`
      case 'GitLab':
        return `<svg viewBox="0 0 24 24" class="${iconClass}" fill="currentColor">
          <path d="M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 0 1-.3-.94l1.22-3.78 2.44-7.51A.42.42 0 0 1 4.82 2a.43.43 0 0 1 .58 0 .42.42 0 0 1 .11.18l2.44 7.49h8.1l2.44-7.51A.42.42 0 0 1 18.6 2a.43.43 0 0 1 .58 0 .42.42 0 0 1 .11.18l2.44 7.51L23 13.45a.84.84 0 0 1-.35.94z"/>
        </svg>`
      case 'Bitbucket':
        return `<svg viewBox="0 0 24 24" class="${iconClass}" fill="currentColor">
          <path d="M.778 1.213a.768.768 0 00-.768.892l3.263 19.81c.084.5.515.868 1.022.873H19.95a.772.772 0 00.77-.646l3.27-20.03a.768.768 0 00-.768-.891zM14.52 15.53H9.522L8.17 8.466h7.561z"/>
        </svg>`
      case 'Unity':
        return `<svg viewBox="0 0 32 32" class="${iconClass}" fill="currentColor">
          <path d="M25.94 25.061l-5.382-9.06 5.382-9.064 2.598 9.062-2.599 9.06zM13.946 24.191l-6.768-6.717h10.759l5.38 9.061-9.372-2.342zM13.946 7.809l9.371-2.342-5.379 9.061h-10.761zM30.996 12.917l-3.282-11.913-12.251 3.193-1.812 3.112-3.68-.027-8.966 8.719 8.967 8.72 3.678-.029 1.817 3.112 12.246 3.192 3.283-11.908-1.864-3.087z"/>
        </svg>
        `
      default:
        return `<svg viewBox="0 0 24 24" class="${iconClass}" fill="currentColor">
          <path d="M12 2L2 7l10 5 10-5-10-5z"/>
        </svg>`
    }
  }

  onMount(() => {
    const savedTheme = (typeof window !== 'undefined' ? localStorage.getItem('theme') : null) as 'light' | 'dark' | 'system' | null
    currentTheme = savedTheme || 'system'
    applyTheme(currentTheme)
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleSystemThemeChange = () => {
      if (currentTheme === 'system') {
        applyTheme('system')
      }
    }
    mediaQuery.addEventListener('change', handleSystemThemeChange)
    return () => {
      mediaQuery.removeEventListener('change', handleSystemThemeChange)
      checkAuthStatus()
    }
  })

  // Reactive statement to save theme
  $: if (typeof window !== 'undefined') {
    localStorage.setItem('theme', currentTheme)
  }
</script>

<!-- Progress bar -->
<div class="fixed top-0 left-0 right-0 h-1 z-50 bg-[#CFDBD5]">
  <div 
    class="h-full transition-all duration-300 ease-in-out bg-[#195B5E]"
    style="width: {((currentPage + 1) / pages.length) * 100}%"
  ></div>
</div>

<button 
  class="fixed top-4 right-4 p-3 rounded-lg bg-white/10 hover:bg-white/20 transition-all duration-200 z-40 backdrop-blur-sm border border-white/20"
  on:click={toggleTheme}
  title="Toggle theme: {currentTheme}"
>
  {@html getThemeIcon()}
</button>

<main class="max-w-7xl mx-auto p-8 flex items-center justify-center min-h-screen">
  {#if !currentUser}
    <div class="rounded-xl p-12 transition-all duration-300 ease-in-out {isTransitioning ? 'opacity-70 translate-x-5' : 'opacity-100 translate-x-0'}">
      <div class="flex flex-col items-center gap-6 mb-8">
        <img 
          alt="App logo" 
          src="{isDarkMode ? logoSecondary : logo}"
          class="w-32 rounded-xl flex-shrink-0"
          
        >
        <div class="text-xl font-semibold text-center flex-1 leading-relaxed text-gray-700 dark:text-gray-100">
          {pages[currentPage].title}
        </div>
      </div>
      
      {#if currentPage === 0}
        <!-- Page 1: Intent -->
        <div class="flex flex-col gap-8">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-2xl mx-auto">
            {#each pages[0].options as option}
              <button 
                on:click={() => toggleOption(option)}
                class="flex items-center justify-center text-left p-6 rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 border-2 border-transparent min-h-[100px] min-w-[220px] hover:-translate-y-0.5 hover:shadow-md
                {selectedOptions.includes(option) 
                  ? 'bg-[#195B5E] text-white' 
                  : 'bg-[#CFDBD5] text-gray-800 hover:bg-[#b8c9be]'}"
              >
                <span class="block w-full text-left break-words">{option}</span>
              </button>
            {/each}
          </div>
          <div class="flex justify-end mt-4">
            <button 
              on:click={nextPage}
              disabled={!canProceed}
              class="px-8 py-3 border-none rounded-xl font-semibold text-base transition-all duration-300 
              {canProceed 
                ? 'bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d] cursor-pointer' 
                : 'bg-[#CFDBD5] text-gray-500 cursor-not-allowed'}"
            >
              Next
            </button>
          </div>
        </div>
      {:else if currentPage === 1}
        <!-- Page 2: Discovery -->
        <div class="flex flex-col gap-8">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-2xl mx-auto">
            {#each pages[1].options as option}
              <button 
                on:click={() => toggleDiscovery(option)}
                class="flex items-center justify-center text-left p-6 rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 border-2 border-transparent min-h-[60px] min-w-[220px] hover:-translate-y-0.5 hover:shadow-md
                {discoveryOptions.includes(option) 
                  ? 'bg-[#195B5E] text-white' 
                  : 'bg-[#CFDBD5] text-gray-800 hover:bg-[#b8c9be]'}"
              >
                <span class="block w-full text-left break-words">{option}</span>
              </button>
            {/each}
          </div>
          {#if discoveryOptions.includes('Other')}
            <div class="flex flex-col items-start max-w-2xl mx-auto w-full">
              <label class="mb-2 text-base font-semibold text-gray-700 dark:text-gray-100" for="otherDiscovery">Please specify:</label>
              <input 
                id="otherDiscovery" 
                type="text" 
                bind:value={discoveryOther} 
                class="w-full p-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-[#195B5E] text-gray-700 dark:text-gray-100 dark:bg-gray-800 dark:border-gray-600" 
                placeholder="How did you find dployr?" 
              />
            </div>
          {/if}
          <div class="flex justify-between mt-4">
            <button 
              on:click={previousPage}
              class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]"
            >
              Back
            </button>
            <button 
              on:click={nextPage}
              disabled={!canProceed}
              class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 
              {canProceed 
                ? 'bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]' 
                : 'bg-[#CFDBD5] text-gray-500 cursor-not-allowed'}"
            >
              Next
            </button>
          </div>
        </div>
      {:else if currentPage === 2}
        <!-- Page 3: App Stage -->
        <div class="flex flex-col gap-8">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-md mx-auto">
            {#each pages[2].options as option}
              <button 
                on:click={() => selectAppStage(option)}
                class="flex items-center justify-center text-left p-6 rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 border-2 border-transparent min-h-[60px] min-w-[180px] hover:-translate-y-0.5 hover:shadow-md
                {appStage === option 
                  ? 'bg-[#195B5E] text-white' 
                  : 'bg-[#CFDBD5] text-gray-800 hover:bg-[#b8c9be]'}"
              >
                <span class="block w-full text-left break-words">{option}</span>
              </button>
            {/each}
          </div>
          <div class="flex justify-between mt-4">
            <button 
              on:click={previousPage}
              class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]"
            >
              Back
            </button>
            <button 
              on:click={nextPage}
              disabled={!canProceed}
              class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 
              {canProceed 
                ? 'bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]' 
                : 'bg-[#CFDBD5] text-gray-500 cursor-not-allowed'}"
            >
              Next
            </button>
          </div>
        </div>
      {:else if currentPage === 3}
        <!-- Page 4: Sign In -->
        <div class="flex flex-col gap-8">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-md mx-auto">
            {#each pages[3].options as option}
              <button 
                on:click={() => handleSignIn(option.toLowerCase())}
                class="flex items-center justify-center p-6 rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 border-2 border-transparent min-h-[60px] min-w-[180px] hover:-translate-y-0.5 hover:shadow-md
                {signInProvider === option 
                  ? 'bg-[#195B5E] text-white' 
                  : 'bg-[#CFDBD5] text-gray-800 hover:bg-[#b8c9be]'}"
              >
                <div class="flex items-center justify-center w-full">
                  {@html getProviderIcon(option)}
                  <span class="break-words">{option}</span>
                </div>
              </button>
            {/each}
          </div>
          <div class="flex justify-between mt-4">
            <button 
              on:click={previousPage}
              class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]"
            >
              Back
            </button>
            <button 
              class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-not-allowed bg-[#CFDBD5] text-gray-500 opacity-60"
              disabled
            >
              Continue
            </button>
          </div>
        </div>
      {/if}
    </div>
  {:else if isAuthenticating} 
    <div class="flex items-center justify-center min-h-[60px]">
      <span class="text-gray-500">Authenticating...</span>
    </div>
  {:else}
    <div class="flex flex-col items-center gap-6">
      <img 
        alt="App logo" 
        src="{isDarkMode ? logoSecondary : logo}"
        class="w-32 rounded-xl flex-shrink-0"
      >
      <div class="text-xl font-semibold text-center leading-relaxed text-gray-700 dark:text-gray-100">
        Welcome back, {currentUser.name.split(' ')[0]}!
      </div>
      <button 
        on:click={handleSignOut}
        class="px-8 py-3 border-none rounded-xl font-semibold text-base bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d] transition-all duration-300"
      >
        Sign Out
      </button>
    </div>
  {/if}
</main>
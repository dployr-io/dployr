<script lang="ts">
  import logo from './assets/images/logo.png'
  import logoSecondary from './assets/images/logo-secondary.png'
  import icon from './assets/images/icon.svg'
  import iconSecondary from './assets/images/icon-secondary.svg'
  import {SignIn, GetCurrentUser, SignOut, GetDeployments, GetProjects} from '../wailsjs/go/main/App.js'
  import { onMount } from 'svelte'
  import { main } from '../wailsjs/go/models';

  let currentTheme: 'light' | 'dark' | 'system' = 'system'
  let isDarkMode = false
  
  $: themeIcon = (() => {
    if (currentTheme === 'system') {
      return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 17.25v1.007a3 3 0 0 1-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0 1 15 18.257V17.25m6-12V15a2.25 2.25 0 0 1-2.25 2.25H5.25A2.25 2.25 0 0 1 3 15V5.25m18 0A2.25 2.25 0 0 0 18.75 3H5.25A2.25 2.25 0 0 0 3 5.25m18 0V12a2.25 2.25 0 0 1-2.25 2.25H5.25A2.25 2.25 0 0 1 3 12V5.25" />
      </svg>
      `
    } else if (currentTheme === 'dark') {
      return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
      </svg>
      `
    } else {
      return `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z" />
      </svg>
      `
    }
  })()
  let currentUser: main.User | null = new main.User({
    ID: '',
    Name: 'John Doe',
    Email: 'john.doe@example.com',
    AvatarURL: '',
    Provider: ''
  });
  let isAuthenticating = false
  let selectedOptions: string[] = []
  let discoveryOptions: string[] = []
  let discoveryOther: string = ''
  let appStage: string = ''
  let signInProvider: string = ''
  let currentPage = 0
  let isTransitioning = false
  let viewMode: 'grid' | 'list' = 'grid'
  let sidebarWidth = 640
  let isResizing = false
  let showFilterDropdown = false
  let selectedSection: string = 'Deployments'
  let appState: { 
    selectedSection: string; 
    selectedProjectId: string | null 
    selectedDeploymentId: string | null
  } = {
    selectedSection: 'Deployments',
    selectedProjectId: null,
    selectedDeploymentId: null
  }
  let projects: main.Project [] = []
  let deployments: main.Deployment[] = []

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

  async function getDeployments() {
    deployments = await GetDeployments()
  }

  async function getProjects() {
    projects = await GetProjects()
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

  function toggleViewMode(mode: 'grid' | 'list') {
    viewMode = mode
  }

  function toggleFilter() {
    showFilterDropdown = !showFilterDropdown
  }

  function selectSection(section: string) {
    selectedSection = section
  }

  function startResize(e: MouseEvent) {
    isResizing = true
    e.preventDefault()
    
    function handleMouseMove(e: MouseEvent) {
      if (!isResizing) return
      const newWidth = Math.max(600, Math.min(800, e.clientX))
      sidebarWidth = newWidth
    }
    
    function handleMouseUp() {
      isResizing = false
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
    
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }

  function parseDuration(int: number): string {
    const minutes = Math.floor(int / 60)
    const seconds = int % 60
    return `${minutes}m ${seconds}s`
  }

  function parseDate(int: number, user: main.User | undefined): string {
    const d = new Date(int)
    const formattedDate = d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    return `${formattedDate} by ${user?.name}`
  }

  function formatProjectDate(date: any): string {    
    if (!date) return 'Unknown date'
    
    const d = new Date(date)
    
    const now = new Date()
    const diffInMs = now.getTime() - d.getTime()
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))
    
    if (diffInDays === 0) {
      return 'Today'
    } else if (diffInDays === 1) {
      return 'Yesterday'
    } else if (diffInDays < 7) {
      return `${diffInDays} days ago`
    } else if (diffInDays < 30) {
      const weeks = Math.floor(diffInDays / 7)
      return `${weeks} ${weeks === 1 ? 'week' : 'weeks'} ago`
    } else if (diffInDays < 365) {
      const months = Math.floor(diffInDays / 30)
      return `${months} ${months === 1 ? 'month' : 'months'} ago`
    } else {
      const years = Math.floor(diffInDays / 365)
      return `${years} ${years === 1 ? 'year' : 'years'} ago`
    }
  }

  function gitIcon(provider: string): string {
    switch (provider) {
      case 'github':
        return `
        <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-brand-github"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M9 19c-4.3 1.4 -4.3 -2.5 -6 -3m12 5v-3.5c0 -1 .1 -1.4 -.5 -2c2.8 -.3 5.5 -1.4 5.5 -6a4.6 4.6 0 0 0 -1.3 -3.2a4.2 4.2 0 0 0 -.1 -3.2s-1.1 -.3 -3.5 1.3a12.3 12.3 0 0 0 -6.2 0c-2.4 -1.6 -3.5 -1.3 -3.5 -1.3a4.2 4.2 0 0 0 -.1 3.2a4.6 4.6 0 0 0 -1.3 3.2c0 4.6 2.7 5.7 5.5 6c-.6 .6 -.6 1.2 -.5 2v3.5" /></svg>
        `
      case 'gitlab':
        return `
        <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-brand-gitlab"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M21 14l-9 7l-9 -7l3 -11l3 7h6l3 -7z" /></svg>
        `
      case 'bitbucket':
        return `
      <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-brand-bitbucket"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3.648 4a.64 .64 0 0 0 -.64 .744l3.14 14.528c.07 .417 .43 .724 .852 .728h10a.644 .644 0 0 0 .642 -.539l3.35 -14.71a.641 .641 0 0 0 -.64 -.744l-16.704 -.007z" /><path d="M14 15h-4l-1 -6h6z" /></svg>
      `
      case 'unity':
        return`
      <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-brand-unity"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M14 3l6 4v7" /><path d="M18 17l-6 4l-6 -4" /><path d="M4 14v-7l6 -4" /><path d="M4 7l8 5v9" /><path d="M20 7l-8 5" /></svg>
      `
      default:
        return `
      <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-brand-git"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M16 12m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M12 8m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M12 16m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0" /><path d="M12 15v-6" /><path d="M15 11l-2 -2" /><path d="M11 7l-1.9 -1.9" /><path d="M13.446 2.6l7.955 7.954a2.045 2.045 0 0 1 0 2.892l-7.955 7.955a2.045 2.045 0 0 1 -2.892 0l-7.955 -7.955a2.045 2.045 0 0 1 0 -2.892l7.955 -7.955a2.045 2.045 0 0 1 2.892 0z" /></svg>  
      `
    }
  }
  // Calculate grid columns based on available width
  $: mainContentWidth = typeof window !== 'undefined' ? window.innerWidth - sidebarWidth - 48 : 800 // 48px for padding
  $: gridCols = Math.max(1, Math.floor(mainContentWidth / 350)) // 350px min card width

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
    checkAuthStatus()
    getDeployments()
    getProjects()
    return () => {
      mediaQuery.removeEventListener('change', handleSystemThemeChange)
    }
  })

  $: if (typeof window !== 'undefined') {
    localStorage.setItem('theme', currentTheme)
  }

  $: if (projects.length > 0 && !selectedProject) {
    selectedProject = projects[0]
  }

  let showProjectDropdown = false
  let showAccountDropdown = false

  let accounts = [
    { name: "zeipo-ai", avatar: "" },
    { name: "username", avatar: "" }
  ]

  let selectedProject = projects[0]
  let selectedAccount = accounts[0]

  function selectProject(project: main.Project) {
    selectedProject = project
    appState.selectedProjectId = project.name 
  }

  function selectAccount(account: { name: string; avatar: string; }) {
    selectedAccount = account
    showAccountDropdown = false
  }

  function selectDeployment(deployment: main.Deployment) {
    appState.selectedDeploymentId = deployment.id
  }

  function toggleProjectDropdown() {
    showProjectDropdown = !showProjectDropdown
    showAccountDropdown = false
  }

  function toggleAccountDropdown(e: { stopPropagation: () => void; }) {
    e.stopPropagation()
    showAccountDropdown = !showAccountDropdown
    showProjectDropdown = false
  }

  // Hide dropdowns on outside click
  if (typeof window !== 'undefined') {
    window.addEventListener('click', () => {
      showProjectDropdown = false
      showAccountDropdown = false
      showFilterDropdown = false
    })
  }
</script>

<!-- Progress bar -->
{#if !!currentUser}
<div class="fixed top-0 left-0 right-0 h-1 z-50 bg-[#CFDBD5]">
  <div 
    class="h-full transition-all duration-300 ease-in-out bg-[#195B5E]"
    style="width: {((currentPage + 1) / pages.length) * 100}%"
  ></div>
</div>
{/if}

<!-- <button 
  class="fixed top-4 right-4 p-3 rounded-lg bg-white/10 hover:bg-white/20 transition-all duration-200 z-40 backdrop-blur-sm border border-white/20"
  on:click={toggleTheme}
  title="Toggle theme: {currentTheme}"
>
<svg viewBox="0 0 24 24" class="w-5 h-5">
  {@html getThemeIcon()}
</svg>
</button> -->

<main class="w-full flex items-center justify-center min-h-screen">
  {#if !!currentUser}
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
  {:else}
    <div class="w-full h-screen flex flex-col">
      <nav class="border-b border-gray-700 flex-shrink-0">
  <div class="relative flex items-center justify-between px-6 py-3">
    <!-- Left: Logo and Project/Account Dropdown -->
    <div class="flex items-center space-x-3 relative flex-shrink-0 min-w-0">
      <img 
        alt="App logo" 
        src="{isDarkMode ? iconSecondary : icon}"
        class="w-8 h-8 rounded-xl flex-shrink-0"
      />

      <!-- Project Dropdown Trigger -->
      <button 
        class="text-white font-semibold flex items-center space-x-1 focus:outline-none min-w-0"
        on:click|stopPropagation={toggleProjectDropdown}
      >
        <span class="text-gray-600 dark:text-gray-200 truncate min-w-0">John's workspace</span>
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 text-gray-600 dark:text-gray-200 flex-shrink-0">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 15 12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9" />
        </svg>
      </button>

      <!-- Project Dropdown -->
      {#if showProjectDropdown}
        <div class="absolute left-0 top-0 mt-9 z-50 card rounded-lg shadow-lg min-w-[220px]">
          <div class="p-2">
            <div class="text-sm font-semibol px-2 py-1 w-fit text-left">{'Projects'}</div>
            {#each projects as project}
              <button 
                class="flex items-center w-full px-3 py-2 rounded app-button-ghost text-left"
                on:click={() => selectProject(project)}
              >
                <div class="w-7 h-7 rounded bg-gray-200 dark:bg-gray-700 flex items-center justify-center mr-2 font-bold">{project.icon}</div>
                <div>
                  <div class="text-gray-600 dark:text-gray-200 font-medium">{project.name}</div>
                </div>
              </button>
            {/each}
            <div class="border-t border-gray-200 dark:border-gray-700 my-2"></div>
              <div class="flex justify-center">
              <button class="w-full max-w-[180px] px-3 py-2 text-left text-blue-600 hover:underline">+ New Project</button>
              </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-700"></div>
          <div class="p-2">
            <button 
              class="flex items-center w-full px-3 py-2 rounded app-button-ghost text-left"
              on:click|stopPropagation={toggleAccountDropdown}
            >
              <span class="font-semibold text-gray-600 dark:text-gray-200">Switch Account</span>
              <svg class="w-4 h-4 ml-auto text-gray-600 dark:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>
            {#if showAccountDropdown}
              <div class="absolute left-full top-0 ml-2 z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg min-w-[180px]">
                <div class="p-2">
                  <div class="text-xs text-gray-500 px-2 py-1">Accounts</div>
                  {#each accounts as account}
                    <button 
                      class="flex items-center w-full px-3 py-2 rounded app-button-ghost text-left"
                      on:click={() => selectAccount(account)}
                    >
                      <div class="w-6 h-6 rounded-full bg-gray-300 dark:bg-gray-600 mr-2"></div>
                      <span>{account.name}</span>
                    </button>
                  {/each}
                  <div class="border-t border-gray-200 dark:border-gray-700 my-2"></div>
                  <button class="w-full px-3 py-2 text-left text-blue-600 hover:underline">+ New Account</button>
                </div>
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </div>

    <!-- Center: Navigation Tabs (Absolutely Centered) -->
    <div class="absolute left-1/2 transform -translate-x-1/2 flex items-center">
      <div class="flex items-center space-x-3 overflow-hidden">
        <a href="#" class="nav-tab active whitespace-nowrap">Overview</a>
        <a href="#" class="nav-tab whitespace-nowrap">Deployments</a>
        <a href="#" class="nav-tab whitespace-nowrap">Resources</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden sm:block">Domains</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden md:block">Insights</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden lg:block">Console</a>
        <a href="#" class="nav-tab whitespace-nowrap hidden xl:block">Settings</a>
      </div>
    </div>

    <!-- Right: User Controls -->
    <div class="flex items-center space-x-3 flex-shrink-0">
      <img src="https://picsum.photos/200/20" alt="Profile" class="h-7 w-7 rounded-full" />
      <div>
        |
      </div>
      <button 
        class="flex w-8 h-8 rounded-lg dark:bg-white/10 bg-white/40 hover:bg-white/20 items-center justify-center border border-white/20"
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0" />
        </svg>
      </button>
      <button 
        class="flex w-8 h-8 rounded-lg dark:bg-white/10 bg-white/40 hover:bg-white/20 items-center justify-center border border-white/20"
        on:click={toggleTheme}
        title="Toggle theme: {currentTheme}"
      >
        {@html themeIcon}
      </button>
    </div>
  </div>
</nav>

      <!-- Sub Header -->
      <div class="relative border-b border-gray-700 px-6 py-3">
  <div class="flex items-center justify-between gap-4">
    <!-- Centered search with filter button -->
    <div class="absolute inset-x-0 top-1/2 transform -translate-y-1/2 flex justify-center pointer-events-none">
      <div class="relative flex items-center w-full max-w-lg pointer-events-auto">
        <!-- Search Input -->
        <div class="relative flex-1">
          <input
            type="text"
            placeholder="Search Projects…"
            class="app-input w-full pl-8 pr-4 py-1.5 text-sm rounded-lg outline-none transition-all"
          />
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
            </svg>
          </span>
        </div>
        
        <!-- Filter Button - positioned right of search -->
        <div class="relative ml-2">
          <button 
            class="p-2 rounded-lg transition-colors flex items-center "
            class:bg-gray-300={showFilterDropdown}
            class:dark:bg-gray-200={showFilterDropdown}
            class:text-gray-500={showFilterDropdown}
            class:dark:text-gray-500={showFilterDropdown}
            class:dark:text-gray-400={showFilterDropdown}
            on:click|stopPropagation={toggleFilter}
            aria-label="Filter"
          >
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75" />
            </svg>
          </button>
          
          <!-- Filter Dropdown -->
          {#if showFilterDropdown}
            <div class="absolute top-full right-0 mt-2 z-200 card rounded-lg shadow-lg min-w-[200px]">
              <div class="p-4">
                <h3 class="font-semibold mb-3">Filter Projects</h3>
                <div class="space-y-2">
                  <label class="flex items-center">
                    <input type="checkbox" class="mr-2" /> Active Projects
                  </label>
                  <label class="flex items-center">
                    <input type="checkbox" class="mr-2" /> Failed Builds
                  </label>
                  <label class="flex items-center">
                    <input type="checkbox" class="mr-2" /> Recent Updates
                  </label>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>

    <!-- Right: Controls (right-aligned, min space) -->
    <div class="flex items-center space-x-2 ml-auto">
      <!-- Grid View Button -->
      <button 
        class="p-2 rounded-lg transition-colors flex items-center"
        class:bg-gray-300={viewMode === 'grid'}
        class:dark:bg-gray-200={viewMode === 'grid'}
        class:text-gray-500={viewMode === 'grid'}
        class:dark:text-gray-500={viewMode === 'grid'}
        class:dark:text-gray-400={viewMode === 'grid'}
        class:hover:bg-gray-100={viewMode !== 'grid'}
        class:dark:hover:bg-gray-800={viewMode !== 'grid'}
        on:click={() => toggleViewMode('grid')}
        aria-label="Grid view"
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 16.875h3.375m0 0h3.375m-3.375 0V13.5m0 3.375v3.375M6 10.5h2.25a2.25 2.25 0 0 0 2.25-2.25V6a2.25 2.25 0 0 0-2.25-2.25H6A2.25 2.25 0 0 0 3.75 6v2.25A2.25 2.25 0 0 0 6 10.5Zm0 9.75h2.25A2.25 2.25 0 0 0 10.5 18v-2.25a2.25 2.25 0 0 0-2.25-2.25H6a2.25 2.25 0 0 0-2.25 2.25V18A2.25 2.25 0 0 0 6 20.25Zm9.75-9.75H18a2.25 2.25 0 0 0 2.25-2.25V6A2.25 2.25 0 0 0 18 3.75h-2.25A2.25 2.25 0 0 0 13.5 6v2.25a2.25 2.25 0 0 0 2.25 2.25Z" />
        </svg>
      </button>
      
      <!-- List View Button -->
      <button 
        class="p-2 rounded-lg transition-colors flex items-center"
        class:bg-gray-300={viewMode === 'list'}
        class:dark:bg-gray-200={viewMode === 'list'}
        class:text-gray-500={viewMode === 'list'}
        class:dark:text-gray-500={viewMode === 'list'}
        class:dark:text-gray-400={viewMode === 'list'}
        class:hover:bg-gray-100={viewMode !== 'list'}
        class:dark:hover:bg-gray-800={viewMode !== 'list'}
        on:click={() => toggleViewMode('list')}
        aria-label="List view"
      >
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
        </svg>
      </button>

      <!-- Add New Button -->
      <button class="app-button px-4 py-2 rounded-lg font-medium flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v6m3-3H9m12 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
        </svg>
        Add New 
      </button>
    </div>
  </div>
      </div>

      <!-- Main Content -->
      <div class="flex flex-1 min-h-0">
        <!-- Left Sidebar -->
        <div class="sidebar border-r border-gray-700 p-6 flex-shrink-0" style="width: {sidebarWidth}px;">
            <!-- Section Content -->
            <div class="min-h-0 flex-1">
                {#if selectedSection === 'Deployments'}
                  <div class="flex flex-col gap-6 h-full">
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
                              <div class="text-sm text-gray-500">{deployments.length}/{deployments.length} deployments</div>
                              <button class="app-button px-4 py-2 rounded-lg">Deploy Now</button>
                          </div>
                          
                          <!-- Deployments List -->
                          <div class="space-y-3">
                            {#each deployments as deployment}
                              <button class="card w-full p-4 rounded-lg border-l-4"
                                   class:border-green-500={deployment.status === 'success'}
                                   class:border-red-500={deployment.status === 'failed'}
                                   class:border-yellow-500={deployment.status === 'pending'}
                                   class:bg-gray-100={appState.selectedDeploymentId === deployment.id}
                                   class:dark:bg-opacity-10={appState.selectedDeploymentId === deployment.id}
                                   on:click={() => selectDeployment(deployment)}>
                                   <div class="flex items-center justify-between mb-2 min-w-0">
                                      <div class="flex items-center space-x-3 min-w-0 flex-1">
                                          <span class="text-xs px-2 py-1 rounded flex-shrink-0"
                                              class:bg-green-100={deployment.status === 'success'}
                                              class:text-green-800={deployment.status === 'success'}
                                              class:bg-red-100={deployment.status === 'failed'}
                                              class:text-red-800={deployment.status === 'failed'}
                                              class:bg-yellow-100={deployment.status === 'pending'}
                                              class:text-yellow-800={deployment.status === 'pending'}
                                          >{(deployment.status)?.toUpperCase()}</span>
                                          <span class="font-mono text-sm truncate">{deployment.id}</span>
                                      </div>
                                      <span class="text-sm text-gray-500 flex-shrink-0 ml-2">{parseDuration(deployment.duration)}</span>
                                  </div>
                                  <div class="flex items-center justify-between text-sm min-w-0">
                                      <div class="flex items-center space-x-2 min-w-0 flex-1">
                                          <span class="flex-shrink-0 flex gap-1">
                                            <svg  xmlns="http://www.w3.org/2000/svg"  width="20"  height="20"  viewBox="0 0 24 24"  fill="none"  stroke="currentColor"  stroke-width="1"  stroke-linecap="round"  stroke-linejoin="round"  class="icon icon-tabler icons-tabler-outline icon-tabler-git-branch"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M7 18m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M7 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M17 6m-2 0a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" /><path d="M7 8l0 8" /><path d="M9 18h6a2 2 0 0 0 2 -2v-5" /><path d="M14 14l3 -3l3 3" /></svg>
                                            {deployment.branch}</span>
                                          <span class="font-mono truncate">{deployment.commitHash}</span>
                                          <span class="truncate">{deployment.message}</span>
                                      </div>
                                      <div class="flex items-center space-x-2 flex-shrink-0 ml-2">
                                          <span class="truncate">{parseDate(deployment.createdAt, deployment.user)}</span>
                                          <div class="w-6 h-6 bg-orange-500 rounded-full flex-shrink-0"></div>
                                      </div>
                                  </div>
                              </button>
                            {/each}
                          </div>
                      </div>
                  </div>
                {:else if selectedSection === 'Logs'}
                  <div class="flex flex-col gap-6 h-full">
                    <!-- Main Content -->
                    <div class="flex-1">
                        <div class="flex items-center justify-between mb-4 gap-2">
                            <input type="text" placeholder="Search for a log entry..." class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" />
                        
                                <button class="app-button-outlined hover:bg-stone-400 hover:dark:bg-stone-600 h-8 w-8 flex items-center justify-center">
                                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
                                  </svg>
                                </button>
                                <button class="app-button-outlined hover:bg-stone-400 hover:dark:bg-stone-600 h-8 w-8 flex items-center justify-center">
                                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
                                  </svg>
                                </button>
                     
                        </div>
                        
                        <!-- Logs Table -->
                        <div class="card rounded-lg overflow-hidden">
                            <div class="grid grid-cols-6 gap-4 p-3 border-stone-400 border-b text-sm font-medium text-left">
                              <div class="text-left">Time</div>
                              <div class="text-left">Status</div>
                              <div class="text-left">Message</div>
                            </div>
                            
                            <div class="">
                              <div class="grid grid-cols-6 gap-4 p-3 border-b border-stone-400 text-sm bg-red-500/5 hover:bg-gray-50 dark:hover:bg-gray-800">
                                <div class="font-mono">14:32:30.22</div>
                                <div class="text-red-500 font-medium flex items-center">
                                  <span class="inline-block w-12 text-left whitespace-nowrap">POST 404</span>
                            
                                </div>
                                <div class="font-mono text-red-500 whitespace-nowrap">The requested resource was not found</div>
                              </div>
                              
                              <div class="grid grid-cols-6 gap-4 p-3 border-b border-stone-400 text-sm hover:bg-gray-50 dark:hover:bg-gray-800">
                                <div class="font-mono">14:32:24.10</div>
                                <div class="text-green-600 font-medium flex items-center">
                                  <span class="inline-block w-12 text-left whitespace-nowrap">GET 200</span>
                          
                                </div>
                                <div class="font-mono whitespace-nowrap">Successful sign-in</div>
                              </div>
                              
                              <div class="grid grid-cols-6 gap-4 p-3 border-b border-stone-400 text-sm bg-orange-500/5 hover:bg-gray-50 dark:hover:bg-gray-800">
                                <div class="font-mono">14:32:24.06</div>
                                <div class="text-orange-600 font-medium flex items-center">
                                  <span class="inline-block w-12 text-left whitespace-nowrap">GET 307</span>

                                </div>
                                <div class="font-mono text-orange-600 whitespace-nowrap">The requested resource was not found</div>
                              </div>
                            </div>
                        </div>
                    </div>
                  </div>
                {:else if selectedSection === 'Resources'}
                    <!-- Resources content -->
                {:else if selectedSection === 'Domains'}
                    <!-- Domains content -->
                {:else if selectedSection === 'Settings'}
                    <!-- Settings content -->
                {:else if selectedSection === 'Insights'}
                    <!-- Insights content -->
                {:else if selectedSection === 'Terminal'}
                    <!-- Terminal content -->
                {/if}
            </div>
        </div>

        <!-- Resize Handle -->
          <div 
            class="w-1 cursor-col-resize transition-colors flex-shrink-0 relative flex items-center justify-center"
            on:mousedown={startResize}
            class:bg-blue-500={isResizing}
          >
            <!-- Capsule Handle Visual Indicator -->
            <div class="absolute left-[-2px] w-1 h-8 bg-gray-400 dark:bg-gray-500 rounded-full hover:bg-gray-500 dark:hover:bg-gray-400 transition-colors"
                class:bg-blue-500={isResizing}
                class:hover:bg-blue-400={isResizing}
                class:w-2={isResizing}>
            </div>
          </div>

          <!-- Main Content Area -->
          <div class="flex-1 p-6 min-h-0 overflow-auto">
            <!-- Project Header with Navigation -->
            <div class="flex items-center justify-between mb-6">
                <h2 class="text-2xl font-bold">{selectedProject.name}</h2>
                
                <!-- Section Navigation Icons -->
                <div class="flex items-center space-x-2">
                    <!-- Deployments -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Deployments'}
                        class:dark:bg-gray-200={selectedSection === 'Deployments'}
                        class:text-gray-500={selectedSection === 'Deployments'}
                        class:dark:text-gray-500={selectedSection === 'Deployments'}
                        class:dark:text-gray-400={selectedSection !== 'Deployments'}
                        class:hover:bg-gray-100={selectedSection !== 'Deployments'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Deployments'}

                        on:click={() => selectSection('Deployments')}
                        title="Deployments"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M6.429 9.75 2.25 12l4.179 2.25m0-4.5 5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0 4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0-5.571 3-5.571-3" />
                        </svg>
                    </button>

                    <!-- Logs -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Logs'}
                        class:dark:bg-gray-200={selectedSection === 'Logs'}
                        class:text-gray-500={selectedSection === 'Logs'}
                        class:dark:text-gray-500={selectedSection === 'Logs'}
                        class:dark:text-gray-400={selectedSection !== 'Logs'}
                        class:hover:bg-gray-100={selectedSection !== 'Logs'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Logs'}
                        on:click={() => selectSection('Logs')}
                        title="Logs"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
                      </svg>
                    </button>

                    <!-- Resources -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Resources'}
                        class:dark:bg-gray-200={selectedSection === 'Resources'}
                        class:text-gray-500={selectedSection === 'Resources'}
                        class:dark:text-gray-500={selectedSection === 'Resources'}
                        class:dark:text-gray-400={selectedSection !== 'Resources'}
                        class:hover:bg-gray-100={selectedSection !== 'Resources'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Resources'}
                        on:click={() => selectSection('Resources')}
                        title="Resources"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 0 1-3-3m3 3a3 3 0 1 0 0 6h13.5a3 3 0 1 0 0-6m-16.5-3a3 3 0 0 1 3-3h13.5a3 3 0 0 1 3 3m-19.5 0a4.5 4.5 0 0 1 .9-2.7L5.737 5.1a3.375 3.375 0 0 1 2.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 0 1 .9 2.7m0 0a3 3 0 0 1-3 3m0 3h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Zm-3 6h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Z" />
                        </svg>
                    </button>

                    <!-- Domains -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Domains'}
                        class:dark:bg-gray-200={selectedSection === 'Domains'}
                        class:text-gray-500={selectedSection === 'Domains'}
                        class:dark:text-gray-500={selectedSection === 'Domains'}
                        class:dark:text-gray-400={selectedSection !== 'Domains'}
                        class:hover:bg-gray-100={selectedSection !== 'Domains'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Domains'}
                        on:click={() => selectSection('Domains')}
                        title="Domains"
                    >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" />
                    </svg>
                    </button>

                    <!-- Settings -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Settings'}
                        class:dark:bg-gray-200={selectedSection === 'Settings'}
                        class:text-gray-500={selectedSection === 'Settings'}
                        class:dark:text-gray-500={selectedSection === 'Settings'}
                        class:dark:text-gray-400={selectedSection !== 'Settings'}
                        class:hover:bg-gray-100={selectedSection !== 'Settings'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Settings'}
                        on:click={() => selectSection('Settings')}
                        title="Settings"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M10.343 3.94c.09-.542.56-.94 1.11-.94h1.093c.55 0 1.02.398 1.11.94l.149.894c.07.424.384.764.78.93.398.164.855.142 1.205-.108l.737-.527a1.125 1.125 0 0 1 1.45.12l.773.774c.39.389.44 1.002.12 1.45l-.527.737c-.25.35-.272.806-.107 1.204.165.397.505.71.93.78l.893.15c.543.09.94.559.94 1.109v1.094c0 .55-.397 1.02-.94 1.11l-.894.149c-.424.07-.764.383-.929.78-.165.398-.143.854.107 1.204l.527.738c.32.447.269 1.06-.12 1.45l-.774.773a1.125 1.125 0 0 1-1.449.12l-.738-.527c-.35-.25-.806-.272-1.203-.107-.398.165-.71.505-.781.929l-.149.894c-.09.542-.56.94-1.11.94h-1.094c-.55 0-1.019-.398-1.11-.94l-.148-.894c-.071-.424-.384-.764-.781-.93-.398-.164-.854-.142-1.204.108l-.738.527c-.447.32-1.06.269-1.45-.12l-.773-.774a1.125 1.125 0 0 1-.12-1.45l.527-.737c.25-.35.272-.806.108-1.204-.165-.397-.506-.71-.93-.78l-.894-.15c-.542-.09-.94-.56-.94-1.109v-1.094c0-.55.398-1.02.94-1.11l.894-.149c.424-.07.765-.383.93-.78.165-.398.143-.854-.108-1.204l-.526-.738a1.125 1.125 0 0 1 .12-1.45l.773-.773a1.125 1.125 0 0 1 1.45-.12l.737.527c.35.25.807.272 1.204.107.397-.165.71-.505.78-.929l.15-.894Z" />
                          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
                        </svg>

                    </button>

                    <!-- Insights -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Insights'}
                        class:dark:bg-gray-200={selectedSection === 'Insights'}
                        class:text-gray-500={selectedSection === 'Insights'}
                        class:dark:text-gray-500={selectedSection === 'Insights'}
                        class:dark:text-gray-400={selectedSection !== 'Insights'}
                        class:hover:bg-gray-100={selectedSection !== 'Insights'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Insights'}
                        on:click={() => selectSection('Insights')}
                        title="Insights"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 0 0 6 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0 1 18 16.5h-2.25m-7.5 0h7.5m-7.5 0-1 3m8.5-3 1 3m0 0 .5 1.5m-.5-1.5h-9.5m0 0-.5 1.5m.75-9 3-3 2.148 2.148A12.061 12.061 0 0 1 16.5 7.605" />
                        </svg>

                    </button>

                    <!-- Terminal -->
                    <button 
                        class="p-2 rounded-lg transition-colors flex items-center gap-2"
                        class:bg-gray-300={selectedSection === 'Terminal'}
                        class:dark:bg-gray-200={selectedSection === 'Terminal'}
                        class:text-gray-500={selectedSection === 'Terminal'}
                        class:dark:text-gray-500={selectedSection === 'Terminal'}
                        class:dark:text-gray-400={selectedSection !== 'Terminal'}
                        class:hover:bg-gray-100={selectedSection !== 'Terminal'}
                        class:dark:hover:bg-gray-800={selectedSection !== 'Terminal'}
                        on:click={() => selectSection('Terminal')}
                        title="Terminal"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
                          <path stroke-linecap="round" stroke-linejoin="round" d="m6.75 7.5 3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 18V6a2.25 2.25 0 0 0-2.25-2.25H5.25A2.25 2.25 0 0 0 3 6v12a2.25 2.25 0 0 0 2.25 2.25Z" />
                        </svg>

                    </button>
                </div>
            </div>

            <!-- Projects Display -->
{#if viewMode === 'grid'}
  <!-- Grid View -->
  <div class="grid gap-6" style="grid-template-columns: repeat({Math.min(gridCols, 3)}, minmax(300px, 1fr));">
    {#each projects as project}
      <button type="button"
        class="card p-6 rounded-lg cursor-pointer transition-all duration-200 text-left relative"  
        class:bg-gray-100={appState.selectedProjectId === project.name}
        class:dark:bg-opacity-10={appState.selectedProjectId === project.name}
        on:click={() => selectProject(project)}
        tabindex="0"
        aria-pressed={appState.selectedProjectId === project.name}
      >
        <button class="absolute top-4 right-4 dark:text-gray-400 text-gray-600 font-medium hover:text-white" tabindex="-1">⋯</button>
        <div class="pr-8 mb-4">
            <div class="flex items-center space-x-3">
                <img src={project.icon} alt={project.name} class="w-10 h-10 rounded" />
                <div class="min-w-0 flex-1">
                    <h3 class="font-semibold truncate">{project.name}</h3>
                    <p class="text-sm dark:text-gray-400 font-medium text-gray-600 truncate">{project.description}</p>
                </div>
            </div>
        </div>

        <div class="flex items-center">
            <div class="flex items-center space-x-2 min-w-0">
                {@html gitIcon(project.provider)}
                <span class="text-sm dark:text-gray-400 text-gray-600 font-medium truncate">{project.url}</span>
            </div>
            <div class="flex ml-auto text-sm dark:text-gray-400 font-medium text-gray-600 flex-shrink-0">
                {formatProjectDate(project.date)}
            </div>
        </div>
      </button>
    {/each}

    <!-- Add New Project Card -->
    <button type="button" 
      class="card p-6 rounded-lg cursor-pointer transition-all duration-200 text-left border-dashed"
      on:click={() => {}}
      tabindex="0"
    >
        <div class="flex items-start justify-between mb-4">
            <div class="flex items-center space-x-3">
                <div class="w-10 h-10 rounded dark:bg-gray-700 bg-gray-500 flex items-center justify-center">
                    <span class="text-white text-xl">+</span>
                </div>
                <div class="min-w-0 flex-1">
                    <h3 class="font-semibold dark:text-gray-400 text-gray-600">Create a new project</h3>
                    <p class="text-sm dark:text-gray-400 font-medium text-gray-600">Start building something amazing</p>
                </div>
            </div>
        </div>
        <div class="flex items-center">
            <div class="flex items-center space-x-2 min-w-0">
                <span class="text-sm dark:text-gray-400 text-gray-600 font-medium">Click to get started</span>
            </div>
        </div>
    </button>
  </div>
{:else}
  <!-- List View -->
  <div class="space-y-3">
    {#each projects as project}
      <button type="button"
        class="card p-4 rounded-lg cursor-pointer transition-all duration-200 text-left w-full relative"  
        class:bg-gray-100={appState.selectedProjectId === project.name}
        class:dark:bg-opacity-10={appState.selectedProjectId === project.name}
        on:click={() => selectProject(project)}
        tabindex="0"
        aria-pressed={appState.selectedProjectId === project.name}
      >
        <button class="absolute top-4 right-4 dark:text-gray-400 text-gray-600 font-medium hover:text-white flex-shrink-0" tabindex="-1">⋯</button>
        <div class="pr-8">
            <div class="flex items-center space-x-3 min-w-0 flex-1">
                <img src={project.icon} alt={project.name} class="w-8 h-8 rounded flex-shrink-0" />
                <div class="min-w-0 flex-1">
                    <h3 class="font-semibold truncate">{project.name}</h3>
                    <p class="text-sm dark:text-gray-400 font-medium text-gray-600 truncate">{project.description}</p>
                </div>
            </div>
            <div class="flex items-center space-x-4 min-w-0 flex-1 mt-2">
                <div class="flex items-center space-x-2 min-w-0">
                    {@html gitIcon(project.provider)}
                    <span class="text-sm dark:text-gray-400 font-medium text-gray-600 truncate">{project.url}</span>
                </div>
                <div class="flex ml-auto text-sm dark:text-gray-400 font-medium text-gray-600 flex-shrink-0">
                    {formatProjectDate(project.date)}
                </div>
            </div>
        </div>
      </button>
    {/each}

    <!-- Add New Project Row -->
    <button type="button" 
      class="card p-4 rounded-lg cursor-pointer transition-all duration-200 text-left w-full border-dashed"
      on:click={() => {}}
      tabindex="0"
    >
        <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3 min-w-0 flex-1">
                <div class="w-8 h-8 rounded bg-gray-700 flex items-center justify-center flex-shrink-0">
                    <span class="text-white text-lg">+</span>
                </div>
                <div class="min-w-0 flex-1">
                    <h3 class="font-semibold text-gray-400">Create a new project</h3>
                    <p class="text-sm dark:text-gray-400 font-medium text-gray-600">Start building something amazing</p>
                </div>
            </div>
            <div class="flex items-center space-x-4 min-w-0 flex-1">
                <div class="flex items-center space-x-2 min-w-0">
                    <span class="text-sm dark:text-gray-400 font-medium text-gray-600">Click to get started</span>
                </div>
            </div>
        </div>
    </button>
  </div>
{/if}
        </div>
      </div>
    </div>
  {/if}
</main>
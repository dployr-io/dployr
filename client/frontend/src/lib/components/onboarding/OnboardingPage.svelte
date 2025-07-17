<script lang="ts">
  import { discoveryOther, email, host, name, isLoading, otp, currentUser } from '../../../stores';
  import OtpInput from '../ui/OtpInput.svelte';

  export let currentPage: number;
  export let pages: any[];
  export let toggleOption: (option: string) => void;
  export let toggleDiscovery: (option: string) => void;
  export let selectAppStage: (option: string) => void;
  export let handleSignIn: () => Promise<void>;
  export let handleMagicCode: () => Promise<void>;
  export let nextPage: () => void;
  export let previousPage: () => void;
  export let canProceed: boolean;
  export let selectedOptions: string[];
  export let discoveryOptions: string[];
  export let appStage: string;
</script>

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
          bind:value={$discoveryOther}
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
  <div class="flex flex-col gap-8 w-96 mx-auto">
    <div class="flex flex-col gap-4 w-full">
      <div>
        <label for="hostname" class="block text-sm font-medium text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Server hostname</label>
        <input 
          id="hostname"
          name="hostname"
          type="text" 
          placeholder="888.888.88.888" 
          bind:value={$host} 
          class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
        />
      </div>
      <!-- <div>
        <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Name</label>
        <input 
          id="name"
          name="name"
          type="text" 
          placeholder="John Doe" 
          bind:value={$name} 
          class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
        />
      </div> -->
      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Email address</label>
        <input 
          id="email"
          name="email"
          type="text" 
          placeholder="admin@acme.inc" 
          bind:value={$email} 
          class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
        />
      </div>
      <!-- {#if usePrivateKey}
        <div>
          <label for="public-key" class="block text-sm font-medium text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Public key</label>
          <input 
            id="public-key"
            name="public-key"
            type="file" 
            bind:value={publicKey} 
            class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all font-medium" 
          />
        </div>
      {:else}
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-400 mb-1 px-4 text-left">Password</label>
          <input 
            id="password"
            name="password"
            type="text" 
            placeholder="Enter password of server" 
            bind:value={password} 
            class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
          />
        </div>
      {/if} -->

      <!-- <div class="flex items-center gap-4">
        <label class="relative inline-flex items-center cursor-pointer select-none">
          <input
            type="checkbox"
            class="sr-only peer"
            bind:checked={usePrivateKey}
          />
          <div
            class="w-11 h-6 bg-gray-100 rounded-full transition-colors duration-200 peer-checked:bg-gray-300"
          ></div>
          <div
            class="absolute left-1 top-1 w-4 h-4 bg-slate-500 rounded-full transform transition-transform duration-200 peer-checked:translate-x-5"
          ></div>
        </label>

        <p class="text-sm font-medium text-gray-700 dark:text-gray-400">
          Use public key?
        </p>
      </div> -->
    </div>
    <div class="flex justify-between mt-4">
      <button 
        on:click={previousPage}
        class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]"
      >
        Back
      </button>
      <button 
        on:click={handleSignIn}
        disabled={!canProceed || $isLoading}
        class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 
        {canProceed && !$isLoading
          ? 'bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]' 
          : 'bg-[#CFDBD5] text-gray-500 cursor-not-allowed'}"      
      >
        {#if $isLoading}
          <div class="flex items-center gap-2">
            <div class="w-5 h-5 border-2 border-gray-500 border-t-transparent rounded-full animate-spin mx-auto" />
            Loading
          </div>
        {:else}
          Sign in
        {/if}
      </button>
    </div>
  </div>
{:else if currentPage === 4}
  <div class="flex flex-col gap-8 w-96 mx-auto">
    <div class="flex flex-col gap-4 w-full">
      <div>
        <label for="magic-code" class="block text-sm font-medium text-gray-700 dark:text-gray-400 mb-2 px-4">Enter six (6) digit magic code sent to your email</label>
        <OtpInput />
      </div> 
    </div>
    <div class="flex justify-between mt-4">
      <button 
        on:click={previousPage}
        class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]"
      >
        Back
      </button>
      <button 
        on:click={handleMagicCode}
        disabled={!canProceed || $isLoading}
        class="px-8 py-3 border-none rounded-xl font-semibold text-base cursor-pointer transition-all duration-300 
        {canProceed && !$isLoading
          ? 'bg-[#195B5E] text-white hover:-translate-y-px hover:bg-[#144a4d]' 
          : 'bg-[#CFDBD5] text-gray-500 cursor-not-allowed'}"      
      >
        {#if $isLoading}
          <div class="flex items-center gap-2">
            <div class="w-5 h-5 border-2 border-gray-500 border-t-transparent rounded-full animate-spin mx-auto" />
            Loading
          </div>
        {:else}
          Verify
        {/if}
      </button>
    </div>
  </div>
{/if}

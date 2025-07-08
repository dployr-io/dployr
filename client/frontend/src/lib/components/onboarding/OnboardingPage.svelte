<script lang="ts">
  import { discoveryOther } from '../../../stores';

  export let currentPage: number;
  export let pages: any[];
  export let toggleOption: (option: string) => void;
  export let toggleDiscovery: (option: string) => void;
  export let selectAppStage: (option: string) => void;
  export let handleSignIn: (provider: string) => Promise<void>;
  export let nextPage: () => void;
  export let previousPage: () => void;
  export let canProceed: boolean;
  export let getProviderIcon: (provider: string) => string;
  export let selectedOptions: string[];
  export let discoveryOptions: string[];
  export let appStage: string;
  export let signInProvider: string;
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

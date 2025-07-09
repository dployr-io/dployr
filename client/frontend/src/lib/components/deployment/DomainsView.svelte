<script>
  import { formatProjectDate, parseDate } from "../../../../src/utils";
  import { domains } from "../../../../src/stores";
  import { AddDomain } from "../../../../wailsjs/go/main/App";

  let domain = '';
  let loading = false;


  async function addDomain() {
    loading = true;
    try {
      return await AddDomain(domain, "foo_bar");
    } finally {
      loading = false;
    }
  }
</script>

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between gap-2 flex-shrink-0">
    <input 
      type="text" 
      placeholder="Search for your subdomain..." 
      bind:value={domain} 
      class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
    />
    <button on:click={addDomain} disabled={loading}
      class="app-button-outlined hover:bg-stone-400 hover:dark:bg-stone-600 h-8 w-8 flex items-center justify-center">
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
        <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
      </svg>
    </button>
  </div>
  <div class="flex flex-col gap-3">
    {#each $domains as domain}
      <button class="flex p-4 text-sm rounded-lg justify-between items-center w-full card">
        <div class="flex flex-col items-start gap-2">
          <p class="text-base font-semibold">{domain.subdomain}</p>
          <p class="font-medium ">{domain.provider}</p>
        </div>
        <div class="flex flex-col items-end gap-2">
          <div class="flex gap-1 items-center">
            {#if domain.verified}
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="size-4 text-green-800 dark:text-green-600">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
              </svg>
              <p class="font-semibold text-green-800 dark:text-green-600">
                Verified
              </p>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="size-4 text-orange-500">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
              </svg>
              <p class="font-semibold text-orange-500">
                Pending
              </p>
            {/if}
          </div>
          <div class="flex gap-2">
            <span class="leading-6 truncate">{parseDate(domain.updatedAt, null)}</span>
            <div class="w-6 h-6 bg-gray-500 rounded-full flex-shrink-0"></div>
          </div>
        </div>
      </button>
    {/each}
  </div>
</div>
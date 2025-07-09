<script lang="ts">
  import { logs } from '../../../stores';
  import { formatLogTime } from '../../../utils';
  import type { LogEntry } from '../../../types';

  let searchQuery = '';

  // Debug logs
  $: console.log('Logs in store:', $logs, 'Length:', $logs.length);

  $: filteredLogs = $logs.filter(log => 
    log.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
    log.host.toLowerCase().includes(searchQuery.toLowerCase()) ||
    log.status.toLowerCase().includes(searchQuery.toLowerCase())
  );

  function getLogStyles(log: LogEntry) {
    switch (log.level) {
      case 'error':
        return {
          rowClass: 'bg-red-500/5 hover:bg-gray-50 dark:hover:bg-gray-800',
          statusClass: 'text-red-500',
          messageClass:  'font-mono text-red-500'
        };
      case 'warning':
        return {
          rowClass: 'bg-orange-500/5 hover:bg-gray-50 dark:hover:bg-gray-800',
          statusClass: 'text-orange-600',
          messageClass: 'font-mono text-orange-600'
        };
      case 'success':
        return {
          rowClass: 'hover:bg-gray-50 dark:hover:bg-gray-800',
          statusClass: 'text-green-600',
          messageClass: 'font-mono'
        };
      default: // info
        return {
          rowClass: 'hover:bg-gray-50 dark:hover:bg-gray-800',
          statusClass: 'text-blue-600',
          messageClass: 'font-mono'
        };
    }
  }
</script>

<div class="h-full flex flex-col gap-4">
  <!-- Main Content -->
    <div class="flex items-center justify-between gap-2 flex-shrink-0">
        <input 
          type="text" 
          placeholder="Search for a log entry..." 
          bind:value={searchQuery}
          class="app-input w-full px-4 py-1.5 text-sm rounded-lg outline-none transition-all" 
        />
        <button class="app-button-outlined hover:bg-stone-400 hover:dark:bg-stone-600 h-8 w-8 flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
          </svg>
        </button>
    </div>
    
    <!-- Logs Table -->
    <div class="flex-1 min-h-0 border-[0.5px] dark:border-stone-700 border-stone-300 rounded-lg overflow-hidden flex flex-col">
        <div class="grid grid-cols-12 gap-4 p-3 dark:border-stone-700 border-stone-300 border-b text-sm font-medium text-left">
          <div class="col-span-2 text-left">Time</div>
          <div class="col-span-2 text-left">Status</div>
          <div class="col-span-6 text-left">Message</div>
        </div>
        
        <div class="flex-1 overflow-auto min-h-0">
          {#each filteredLogs as log}
            {@const styles = getLogStyles(log)}
            <div class="grid grid-cols-12 gap-4 p-3 border-b dark:border-stone-700 border-stone-300 text-sm {styles.rowClass}">
              <div class="col-span-2 font-mono whitespace-nowrap truncate">{formatLogTime(log.createdAt)}</div>
              <div class="col-span-2 {styles.statusClass} font-medium flex items-center">
                <span class="inline-block w-16 text-left whitespace-nowrap truncate">{log.status}</span>
              </div>
              <div class="flex col-span-6 {styles.messageClass} whitespace-nowrap truncate" title={log.message}>
                {log.message}
              </div>
            </div>
          {/each}
          
          {#if filteredLogs.length === 0}
            <div class="p-8 text-center text-gray-500">
              {searchQuery ? 'No logs match your search' : 'No logs available'}
            </div>
          {/if}
        </div>
    </div>
</div>

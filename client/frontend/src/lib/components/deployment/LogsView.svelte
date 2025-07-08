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

<?php

namespace App\Jobs\Services;

use App\Models\Remote;
use App\Services\CaddyService;
use App\Services\GitRepoService;
use App\Services\DirectoryService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Log;

class ProcessBlueprint implements ShouldQueue
{
    use Queueable;
    
    public function __construct(
        private array $config = [],
    ) {}

    public function handle(): void
    {
        $serviceName = $this->config['name'] ?? null;
        $remoteId = $this->config['remote'] ?? null;
        $path = 'home/dployr/services';
        
        try {
            $remote = Remote::findOrFail($remoteId);
            DirectoryService::setupFolder($path);
    
            $remoteService = new GitRepoService();
            $remoteService->cloneRepo($remote->name, $remote->repository, $remote->provider, $path);
            
            $caddyFile = CaddyService::getStatus();

            Log::debug($caddyFile);

            Log::info("Successfully created service $serviceName");
        } catch (\RuntimeException $e) {
            Log::error("Runtime exception on service $serviceName: " . $e->getMessage());
        } catch (\Exception $e) {
            $errorMessage = $e instanceof \Throwable ? $e->getMessage() : 'An unexpected error occurred.';
            Log::error("Failed to create service $serviceName: " . $errorMessage);        
        }
    }
}

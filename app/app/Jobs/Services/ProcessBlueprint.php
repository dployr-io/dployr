<?php

namespace App\Jobs\Services;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Log;
use App\Models\Blueprint;
use App\Enums\JobStatus;

class ProcessBlueprint implements ShouldQueue
{
    use Queueable;
    
    public function __construct(
        private array $config = [],
    ) {}

    public function handle(): void
    {
        $name = $this->config['name'] ?? null;

        try {
            Log::info('Successfully created blueprint ' . $name);
        } catch (\Throwable $th) {
            Log::error('Failed to create blueprint '. $name);
        }
    }
}

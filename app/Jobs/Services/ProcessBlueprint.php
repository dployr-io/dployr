<?php

namespace App\Jobs\Services;

use App\Models\Blueprint;
use App\Services\Blueprint\BlueprintService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;

class ProcessBlueprint implements ShouldQueue
{
    use Queueable;

    public function __construct(
        private Blueprint $blueprint,
    ) {}

    public function uniqueId(): string
    {
        return $this->blueprint->id;
    }
    
    public function uniqueFor(): int
    {
        return 3600; // 1 hour
    }

    public function handle(): void
    {
        $blueprintService = new BlueprintService($this->blueprint);
        $blueprintService->process();
    }
}

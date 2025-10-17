<?php

namespace App\Jobs;

use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Process;

class ExecuteCmd implements ShouldQueue
{
    use Queueable;

    public function __construct(
        private string $command,
        private array $options = []
    ) {}

    public function handle(): void
    {
        $process = Process::timeout($this->options['timeout'] ?? 300);

        if (isset($this->options['working_dir'])) {
            $process = $process->path($this->options['working_dir']);
        }

        if (! empty($this->options['environment'])) {
            $process = $process->env($this->options['environment']);
        }

        $result = $process->run($this->command);

        Log::info('Command executed', [
            'command' => $this->command,
            'exit_code' => $result->exitCode(),
            'successful' => $result->successful(),
        ]);

        if ($result->failed()) {
            Log::error('Command failed', [
                'command' => $this->command,
                'error' => $result->errorOutput(),
            ]);
        }
    }
}

<?php

namespace App\Services;

use Illuminate\Support\Facades\Process;
use App\Jobs\ExecuteCmdJob;
use App\DTOs\CmdResult;

class CmdService
{
    /**
     * Execute a command using the process api
     * @param string $command
     * @param array $options
     * @return CmdResult
     */
    public function execute(string $command, array $options = []): CmdResult
    {
        $options = array_merge([
            'timeout' => 300, // 5 minutes default
            'working_directory' => null,
            'environment' => [],
            'async' => false,
        ], $options);

        if ($options['async']) {
            return $this->executeAsync($command, $options);
        }

        return $this->executeSync($command, $options);
    }

    /**
     * Command is executued in the same process. 
     * This results to a blocking request.
     * @param string $command
     * @param array $options
     * @return CmdResult
     */
    private function executeSync(string $command, array $options): CmdResult
    {
        $process = Process::timeout($options['timeout']);
        
        if ($options['working_directory']) {
            $process = $process->path($options['working_directory']);
        }

        if (!empty($options['environment'])) {
            $process = $process->env($options['environment']);
        }

        $result = $process->run($command);

        return new CmdResult(
            command: $command,
            exitCode: $result->exitCode(),
            output: $result->output(),
            errorOutput: $result->errorOutput(),
            successful: $result->successful()
        );
    }

    /**
     * Command is dispached to a background worker.
     * @param string $command
     * @param array $options
     * @return CmdResult
     */
    private function executeAsync(string $command, array $options): CmdResult
    {
        ExecuteCmdJob::dispatch($command, $options);
        
        return new CmdResult(
            command: $command,
            exitCode: null,
            output: 'Command queued for execution',
            errorOutput: '',
            successful: true,
            isAsync: true
        );
    }
}
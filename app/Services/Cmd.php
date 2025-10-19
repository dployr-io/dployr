<?php

namespace App\Services;

use App\DTOs\CmdResult;
use App\Jobs\ExecuteCmd;
use Illuminate\Support\Facades\Process;
use Illuminate\Process\Exceptions\ProcessTimedOutException;
use Log;

final class Cmd
{
    /**
     * Execute a command using the process api
     */
    public static function execute(string $command, array $options = []): CmdResult
    {
        $options = array_merge([
            'timeout' => 300, // 5 minutes default
            'working_dir' => null,
            'environment' => [],
            'async' => false,
        ], $options);

        if ($options['async']) {
            return self::executeAsync($command, $options);
        }

        return self::executeSync($command, $options);
    }

    /**
     * Command is executued in the same process.
     * This results to a blocking request.
     */
    private static function executeSync(string $command, array $options): CmdResult
    {
        try {
            $process = Process::timeout($options['timeout']);

            if ($options['working_dir']) {
                $process = $process->path($options['working_dir']);
            }

            if (! empty($options['environment'])) {
                $process = $process->env($options['environment']);
            }

            $result = $process->run($command);

            if ($result->successful()) {
                $output = $result->output();
                $message = $command;
                if ($output !== null && $output !== '') {
                    $message .= ' => '.$output;
                }
                Log::info($message);
            } else {
                $errorOutput = $result->errorOutput();
                $message = $command;
                if ($errorOutput !== null && $errorOutput !== '') {
                    $message .= ' => '.$errorOutput;
                }
                Log::error($message);
            }

            return new CmdResult(
                command: $command,
                exitCode: $result->exitCode(),
                output: $result->output(),
                errorOutput: $result->errorOutput(),
                successful: $result->successful()
            );
        } catch (ProcessTimedOutException $e) {
            Log::error("Command {$command} timed out");
            return new CmdResult($command, 124, '', 'Timed out', false);
        } catch (\Throwable $e) {
            Log::error("Command failed: {$command} => {$e->getMessage()}");
            return new CmdResult($command, 1, '', $e->getMessage(), false);
        }
    }

    /**
     * Command is dispached to a background worker.
     */
    private static function executeAsync(string $command, array $options): CmdResult
    {
        ExecuteCmd::dispatch($command, $options);

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

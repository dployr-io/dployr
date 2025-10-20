<?php

namespace App\Services;

use App\DTOs\CmdResult;
use Illuminate\Support\Facades\Process;
use Illuminate\Process\Exceptions\ProcessTimedOutException;

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
        ], $options);
        
        try {
            $process = Process::timeout($options['timeout']);

            if ($options['working_dir']) {
                $process = $process->path($options['working_dir']);
            }

            if (! empty($options['environment'])) {
                $process = $process->env($options['environment']);
            }

            $output = '';
            $errorOutput = '';

            $result = $process->run($command, function (string $type, string $buffer) use (&$output, &$errorOutput, $command) {
                if ($type === 'out') {
                    $output .= $buffer;
                } else {
                    $errorOutput .= $buffer;
                }
                
                $logMethod = $type === 'out' ? 'info' : 'error';
                \Log::$logMethod("[$command] → $buffer");
            });

            return new CmdResult(
                command: $command,
                exitCode: $result->exitCode(),
                output: $output,
                errorOutput: $errorOutput,
                successful: $result->successful()
            );
        } catch (ProcessTimedOutException $e) {
            \Log::error("Command {$command} timed out");
            return new CmdResult($command, 124, '', 'Timed out', false);
        } catch (\Throwable $e) {
            \Log::error("Command failed: {$command} => {$e->getMessage()}");
            return new CmdResult($command, 1, '', $e->getMessage(), false);
        }
    }
}

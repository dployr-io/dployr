<?php

namespace App\Services;

class LogStreamService
{
    /**
     * Stream logs from a specified log file using Server-Sent Events (SSE).
     * The log file is read incrementally, sending new lines to the client as they are added.
     * @param string $filePath
     */
    public static function stream(string $filePath)
    {
        $file = storage_path($filePath);
        $position = 0;

        // Disable output buffering
        while (ob_get_level() > 0) ob_end_flush();
        ob_implicit_flush(true);

        echo "retry: 2000\n\n";

        while (true) {
            if (connection_aborted()) {
                break; // stop if the client disconnects
            }

            if (!file_exists($file)) {
                echo "event: error\n";
                echo "data: " . json_encode(['message' => "Log file does not exist: $file"]) . "\n\n";
                flush();
                return; // exit the loop and end the stream
            }

            clearstatcache(false, $file);
            $size = filesize($file);

            // New data?
            if ($size > $position) {
                $handle = fopen($file, 'r');
                fseek($handle, $position);
                while (($line = fgets($handle)) !== false) {
                    $trimmedLine = trim($line);
                    if ($trimmedLine === '') {
                        continue; // skip empty lines
                    }
                    echo "data: " . json_encode(['message' => $trimmedLine]) . "\n\n";
                    flush();
                }
                $position = ftell($handle);
                fclose($handle);
            }

            sleep(1); // refresh every second
        }
    }
}

<?php

namespace App\Http\Controllers\Projects\Services;

use Symfony\Component\HttpFoundation\StreamedResponse;

class LogsController 
{
    /**
     * Stream logs from a deployment
     */
    public function logs(): StreamedResponse
    {
        return new StreamedResponse(function () {
            $logFile = '/home/dployr/storage/logs/laravel.log';
            $lines = array_slice(file($logFile), -100);
            $handle = popen("tail -f -n 0 $logFile", 'r');

            foreach ($lines as $line) {
                echo json_encode(['message' => trim($line)]) . "\n\n";
                ob_flush();
                flush();
            }
            while (!feof($handle)) {
                $line = fgets($handle);
                if ($line) {
                    echo json_encode(['message' => trim($line)]) . "\n\n";
                    ob_flush();
                    flush();
                }
            }
            pclose($handle);
        }, 200, [
            'Content-Type' => 'text/event-stream',
            'Cache-Control' => 'no-cache',
            'X-Accel-Buffering' => 'no',
        ]);
    }
}
?>
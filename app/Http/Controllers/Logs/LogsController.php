<?php

namespace App\Http\Controllers\Logs;

use App\Http\Controllers\Controller;
use App\Services\LogStreamService;
use Symfony\Component\HttpFoundation\StreamedResponse;

class LogsController extends Controller
{
    /**
     * Stream logs from a deployment
     */
    public function stream(): StreamedResponse
    {
        return new StreamedResponse(
            fn () => LogStreamService::stream('logs/laravel.log'),
            200,
            [
                'Content-Type' => 'text/event-stream',
                'Cache-Control' => 'no-cache',
                'X-Accel-Buffering' => 'no',
            ]);
    }
}

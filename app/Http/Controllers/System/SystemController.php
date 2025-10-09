<?php

namespace App\Http\Controllers\System;

use App\Facades\Command;
use Inertia\Controller;

class SystemController extends Controller
{
    /**
     * Show the system properties.
     *
     * @return \Illuminate\Http\JsonResponse
     */
    public function index()
    {
        $diskUsage = Command::execute('df -h');
        $memoryUsage = Command::execute('free -m');

        return response()->json([
            'disk' => $diskUsage->output,
            'memory' => $memoryUsage->output,
            'success' => $diskUsage->successful && $memoryUsage->successful,
        ]);
    }
}

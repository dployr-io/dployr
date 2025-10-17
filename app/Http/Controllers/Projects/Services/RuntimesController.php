<?php

namespace App\Http\Controllers\Projects\Services;

use App\Services\Runtime\RuntimeService;
use Illuminate\Http\JsonResponse;

class RuntimesController
{
    /**
     * List all available runtime versions
     */
    public function list(): JsonResponse
    {
        $param = request()->query('runtime');

        if (! is_string($param)) {
            return response()->json([]);
        }

        $runtime = new RuntimeService($param);

        return response()->json($runtime->list());
    }
}

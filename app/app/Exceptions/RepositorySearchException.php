<?php

namespace App\Exceptions;

use Exception;
use Illuminate\Http\JsonResponse;

class RepositorySearchException extends Exception
{
    /**
     * Render the exception into an HTTP response.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\JsonResponse
     */
    public function render($request): JsonResponse
    {
        return response()->json([
            'error' => 'Failed to search for remote repository'
        ], 502);
    }
}

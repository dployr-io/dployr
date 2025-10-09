<?php

namespace App\Http\Controllers\Projects\Services;

use App\Http\Controllers\Controller;
use App\Models\Blueprint;
use Illuminate\Http\JsonResponse;
use Inertia\Inertia;

class BlueprintsController extends Controller
{
    /**
     * Show deployments page
     */
    public function index()
    {
        return Inertia::render('deployments/index');
    }

    /**
     * Fetch all blueprints
     */
    public function fetch(): JsonResponse
    {
        return response()->json(
            Blueprint::all()->map(fn ($blueprint) => [
                'id' => $blueprint->id,
                'config' => $blueprint->config,
                'status' => $blueprint->status,
                'created_at' => $blueprint->created_at,
                'updated_at' => $blueprint->updated_at,
            ],
            )
        );
    }

    /**
     * Show a deployment
     */
    public function show(Blueprint $blueprint)
    {
        return Inertia::render('deployments/view-deployment', [
            'blueprint' => [
                'id' => $blueprint->id,
                'config' => $blueprint->config,
                'status' => $blueprint->status,
                'created_at' => $blueprint->created_at,
                'updated_at' => $blueprint->updated_at,
            ],
        ]);
    }
}

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
        $query = Blueprint::query();

        if (request()->query('spec')) {
            $query->where('save_spec', true);
        }

        return response()->json(
            $query
                ->get()
                ->map(fn ($blueprint) => [
                    'id' => $blueprint->id,
                    'config' => array_merge(
                        is_array($blueprint->config) ? $blueprint->config : json_decode($blueprint->config, true) ?? [],
                        array_filter([
                            'remote' => $blueprint->remote_obj,
                            'ci_remote' => $blueprint->ci_remote_obj,
                        ], fn ($value) => $value !== null)
                    ),
                    'status' => $blueprint->status,
                    'spec' => $blueprint->spec,
                    'created_at' => $blueprint->created_at,
                    'updated_at' => $blueprint->updated_at,
                ])
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
                'config' => array_merge(
                    is_array($blueprint->config) ? $blueprint->config : json_decode($blueprint->config, true) ?? [],
                    array_filter([
                        'remote' => $blueprint->remote_obj,
                        'ci_remote' => $blueprint->ci_remote_obj,
                    ], fn ($value) => $value !== null)
                ),
                'status' => $blueprint->status,
                'created_at' => $blueprint->created_at,
                'updated_at' => $blueprint->updated_at,
            ],
        ]);
    }
}

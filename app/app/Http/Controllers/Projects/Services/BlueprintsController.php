<?php

namespace App\Http\Controllers\Projects\Services;

use App\Models\Blueprint;
use Inertia\Inertia;
use App\Http\Controllers\Controller;

class BlueprintsController extends Controller 
{
    /**
     * Show deployments page
     */
    public function index()
    {
        return Inertia::render('deployments/index', [
            'blueprints' => Blueprint::all()->map( fn($blueprint) => 
                [
                    'id' => $blueprint->id,
                    'config' => $blueprint->config,
                    'status' => $blueprint->status,
                    'created_at' => $blueprint->created_at,
                    'updated_at' => $blueprint->updated_at,
                ],
            ),
        ]);
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
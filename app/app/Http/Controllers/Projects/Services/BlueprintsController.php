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
        return Inertia::render('projects/services/deployments/index', [
            'blueprints' => Blueprint::all()->map( fn($blueprint) => 
                [
                    'id' => $blueprint->id,
                    'config' => $blueprint->config,
                    'status' => $blueprint->status,
                ],
            ),
        ]);
    }

    /**
     * Show a deployment
     */
    public function show(Blueprint $blueprint)
    {
        return Inertia::render('projects/services/deployments/view-deployment', [
            'blueprint' => [
                'id' => $blueprint->id,
                'config' => $blueprint->config,
                'status' => $blueprint->status,
            ],
        ]);
    }  
}
<?php

namespace App\Http\Controllers\Projects\Services;

use Inertia\Inertia;
use Illuminate\Http\Request;
use Illuminate\Http\RedirectResponse;
use App\Http\Controllers\Controller;
use App\Models\Service;

class ServicesController extends Controller 
{
    /**
     * Show all service creation page.
     */
    public function index()
    {
        return Inertia::render('projects/services/deploy-service');
    }

    /**
     * Show a project's page.
     */
    public function show(Service $project)
    {
        return Inertia::render('projects/services/view-service', [
            'service' => [
                'id' => $project->id,
                'name' => $project->name,
            ],
        ]);
    }

    /**
     * Handle a new project request.
     *
     * @throws \Illuminate\Validation\ValidationException
     */
    public function store(Request $request): RedirectResponse 
    {
        $request->validate([
            'name' => ['required'],
        ]);

        Service::create([
            'name' => $request->name,
            'description' => $request->description,
        ]);

        return back()->with('success', __('Your project was created successfully.'));
    }      
}
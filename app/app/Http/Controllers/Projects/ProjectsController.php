<?php

namespace App\Http\Controllers\Projects;

use Inertia\Inertia;
use Illuminate\Http\Request;
use Illuminate\Http\RedirectResponse;
use App\Http\Controllers\Controller;
use App\Models\Project;
use App\Rules\RemoteRepo;
use App\Services\GitRepoService;

class ProjectsController extends Controller 
{
    /**
     * Show all projects page.
     */
    public function index()
    {
        return Inertia::render('projects/index', [
            'projects' => Project::all()->map( fn($project) => 
                [
                    'id' => $project->id,
                    'name' => $project->name,
                    'description' => $project->description,
                ]
            ),
        ]);
    }

    /**
     * Show a project's page.
     */
    public function show(Project $project)
    {
        return Inertia::render('projects/view-project', [
            'project' => [
                'id' => $project->id,
                'name' => $project->name,
                'description' => $project->description,
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

        Project::create([
            'name' => $request->name,
            'description' => $request->description,
        ]);

        return back()->with('success', __('Your project was created successfully.'));
    }      
}
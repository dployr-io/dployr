<?php

namespace App\Http\Controllers\Projects;

use App\Http\Controllers\Controller;
use App\Models\Project;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;

class ProjectsController extends Controller
{
    /**
     * Show all projects page.
     */
    public function index()
    {
        return Inertia::render('projects/index');
    }

    /**
     * Fetch all projects
     */
    public function fetch(): JsonResponse
    {
        return response()->json(
            Project::all()->map(fn ($project) => [
                'id' => $project->id,
                'name' => $project->name,
                'description' => $project->description,
            ]
            ),
        );
    }

    /**
     * Show a project's page.
     */
    public function show(Project $project)
    {
        return Inertia::render('projects/services/index', [
            'project' => [
                'id' => $project->id,
                'name' => $project->name,
                'description' => $project->description,
            ],
            'services' => $project->services->map(fn ($service) => [
                'id' => $service->id,
                'name' => $service->name,
                'source' => $service->source,
                'runtime' => $service->runtime,
                'run_cmd' => $service->run_cmd,
                'port' => $service->port,
                'working_dir' => $service->working_dir,
                'static_dir' => $service->static_dir,
                'image' => $service->image,
                'spec' => $service->spec,
                'env_vars' => $service->env_vars,
                'secrets' => $service->secrets,
                'remote_id' => $service->remote_id,
                'ci_remote_id' => $service->ci_remote_id,
            ]),
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

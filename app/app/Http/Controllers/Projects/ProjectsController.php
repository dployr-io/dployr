<?php

namespace App\Http\Controllers\Projects;

use Exception;
use Inertia\Inertia;
use Illuminate\Http\Request;
use Illuminate\Http\RedirectResponse;
use App\Http\Controllers\Controller;
use App\Models\Project;
use App\Rules\RemoteRepo;
use App\Services\GitRepository;
use App\Exceptions\RepositorySearchException;

class ProjectsController extends Controller 
{
    /**
     * Show the projects page.
     */
    public function index()
    {
        return Inertia::render('projects/index', [
            'projects' => Project::all()->map( fn($project) => 
                [
                    'id' => $project->id,
                    'name' => $project->name,
                    'remote' => $project->remote,
                    'branch' => $project->branch,
                    'repository' => $project->repository,
                    'lastCommitMessage' => "This is the last commit message"
                ]
            ),
        ]);
    }

    /**
     * Search for a remote repository.
     * 
     * @throws \Illuminate\Validation\ValidationException
     * @throws \App\Exceptions\RepositorySearchException
     */
    public function search(Request $request) 
    {
        $request->validate([
            'remote_repo' => ['required', new RemoteRepo()],
        ]);

        $response = GitRepository::search($request->remote_repo);

        if (!$response->success) 
        {
            throw new RepositorySearchException();
        }

        return back()->withInput()->with('branches', $response->data['branches'] ?? []);
    }

    /**
     * Handle an incoming new project request.
     *
     * @throws \Illuminate\Validation\ValidationException
     */
    public function store(Request $request): RedirectResponse 
    {
        $request->validate([
            'remote_repo' => ['required', new RemoteRepo()],
            'branch' => 'required',
        ]);

        $parsed = GitRepository::parse($request->remote_repo);

        Project::create([
            'name' => $parsed['name'],
            'repository' => $parsed['repository'],
            'remote' => $parsed['remote'],
            'branch' => $request->branch,
        ]);

        return back()->with('status', __('Your project was created, and the import is now in progress.'));
    }      
}
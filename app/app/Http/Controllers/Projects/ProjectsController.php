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

        $repo_service = new GitRepoService();

        try {
            $response = $repo_service->search($request->remote_repo);

            if (!$response->success) {
                return back()->withInput()->with('error', $response->error['message'] ?? 'Failed to fetch repository.');
            }
            return back()->withInput()->with('data', $response->data['branches'] ?? []);
        } catch (\RuntimeException $e) {
            return back()->withInput()->with('error', $e->getMessage());
        } catch (\Exception $e) {
            $errorMessage = $e instanceof \Throwable ? $e->getMessage() : 'An unexpected error occurred.';
            return back()->withInput()->with('error', $errorMessage ?: 'An unexpected error occurred.');
        }
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

        $parsed = GitRepoService::parse($request->remote_repo);

        Project::create([
            'name' => $parsed['name'],
            'repository' => $parsed['repository'],
            'remote' => $parsed['remote'],
            'branch' => $request->branch,
        ]);

        return back()->with('status', __('Your project was created, and the import is now in progress.'));
    }      
}
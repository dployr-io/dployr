<?php

namespace App\Http\Controllers\Projects\Remotes;

use Inertia\Inertia;
use Illuminate\Http\Request;
use Illuminate\Http\RedirectResponse;
use App\Http\Controllers\Controller;
use App\Models\Remote;
use App\Rules\RemoteRepo;
use App\Services\GitRepoService;

class RemotesController extends Controller 
{
    /**
     * Show all remotes page.
     */
    public function index()
    {
        return Inertia::render('projects/remotes/index', [
            'remotes' => Remote::all()->map( fn($remote) => 
                [
                    'id' => $remote->id,
                    'name' => $remote->name,
                    'remote' => $remote->remote,
                    'branch' => $remote->branch,
                    'repository' => $remote->repository,
                    'commit' => $remote->commit,
                ]
            ),
        ]);
    }

    /**
     * Show a remote's page.
     */
    public function show(Remote $remote)
    {
        return Inertia::render('projects/remotes/view-remote', [
            'remote' => [
                'id' => $remote->id,
                'name' => $remote->name,
                'remote' => $remote->remote,
                'branch' => $remote->branch,
                'repository' => $remote->repository,
                'commit' => $remote->commit,
            ],
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
     * Handle a new remote request.
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

        Remote::create([
            'name' => $parsed['name'],
            'repository' => $parsed['repository'],
            'remote' => $parsed['remote'],
            'branch' => $request->branch,
            'commit' => $request->commit,
        ]);

        return back()->with('success', __('Your remote repository was added successfully.'));
    }      
}
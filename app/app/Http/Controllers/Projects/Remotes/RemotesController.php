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
        
        return Inertia::render('resources/remotes', [
            'remotes' => Remote::all()->map( function($remote) { 
                
                try {
                    $repo_service = new GitRepoService();
                    $commit = $repo_service->getLatestCommitMessage($remote->name, $remote->repository, $remote->provider);
                    $commit_message = $commit['message'];
                    $avatar_url = $commit['avatar_url'];
                } catch (\RuntimeException $e) {
                    $commit_message = 'Unable to fetch commit';
                    $avatar_url = null;
                }

                return [
                    'id' => $remote->id,
                    'name' => $remote->name,
                    'provider' => $remote->provider,
                    'branch' => $remote->branch,
                    'repository' => $remote->repository,
                    'commit_message' => $commit_message,
                    'avatar_url' => $avatar_url,
                ];
            }),
        ]);
    }

    /**
     * Show a remote's page.
     */
    public function show(Remote $remote)
    {
        try {
            $repo_service = new GitRepoService();
            $commit = $repo_service->getLatestCommitMessage($remote->name, $remote->repository, $remote->provider);
            $commit_message = $commit['message'];
            $avatar_url = $commit['avatar_url'];
        } catch (\RuntimeException $e) {
            $commit_message = 'Unable to fetch commit';
            $avatar_url = null;
        }

        return Inertia::render('resources/remotes/view-remote', [
            'remote' => [
                'id' => $remote->id,
                'name' => $remote->name,
                'provider' => $remote->provider,
                'branch' => $remote->branch,
                'repository' => $remote->repository,
                'commit_message' => $commit_message,
                'avatar_url' => $avatar_url,
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
  
        try {
            $repo_service = new GitRepoService();
            $response = $repo_service->searchRemote($request->remote_repo);

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
            'provider' => $parsed['provider'],
            'branch' => $request->branch,
        ]);

        return back()->with('success', __('Your remote repository was added successfully.'));
    }      
}
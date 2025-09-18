<?php

namespace App\Services\RemoteProviderServices;

use Http;
use App\DTOs\RemoteRepo;
use App\Enums\RemoteType;
use App\Facades\AppConfig;
use Illuminate\Support\Facades\Log;

class GithubService extends RemoteProviderService {
    public function __construct()
    {
        $token = AppConfig::get('github_token');

        if (empty($token) || !is_string($token)) {
            throw new \RuntimeException("GitHub token required. Check Settings > Configuration.", 1);
        }

        parent::__construct($token);
    }

    public function search(string $name, string $repository): RemoteRepo
    {
        $headers = [
            'Authorization' => "Bearer {$this->token}",
            'Accept'        => 'application/vnd.github.v3+json',
        ];

        $repoUrl = "https://api.github.com/repos/{$name}/{$repository}";
        $repoResponse = Http::withHeaders($headers)->get($repoUrl);

        if (!$repoResponse->successful()) {
            $msg = $repoResponse->json('message') ?? $repoResponse->body();
            throw new \RuntimeException("Failed to fetch repository: {$repoUrl} ({$msg})", 1);
        }

        $repoData = $repoResponse->json();

        $branchesUrl = "{$repoUrl}/branches";
        $branchesResponse = Http::withHeaders($headers)->get($branchesUrl);

        if (!$branchesResponse->successful()) {
            $msg = $branchesResponse->json('message') ?? $branchesResponse->body();
            throw new \RuntimeException("Failed to fetch branches: {$branchesUrl} ({$msg})", 1);
        }

        $branchesData = $branchesResponse->json();
        $branches = array_map(fn($branch) => $branch['name'], $branchesData);

        return new RemoteRepo(
            name: "{$repository}/{$name}",
            remote_type: RemoteType::Github,
            branches: $branches,
            url: $repoData['clone_url'] ?? '',
            avatar_url: $repoData['owner']['avatar_url'] ?? '',
        );
    }
}
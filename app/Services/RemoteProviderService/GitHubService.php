<?php

namespace App\Services\RemoteProviderService;

use App\DTOs\ApiResponse;
use App\Enums\RemoteType;
use App\Facades\AppConfig;
use App\Services\Cmd;
use App\Services\HttpService;

class GitHubService extends RemoteProviderService
{
    public function __construct()
    {
        $token = AppConfig::get('github_token');

        if (empty($token) || ! is_string($token)) {
            throw new \RuntimeException('GitHub token required. Check Settings > Configuration.', 1);
        }

        parent::__construct($token);
    }

    public function search(string $name, string $repository, string $provider): ApiResponse
    {
        if (parent::getRemoteType($provider) != RemoteType::GitHub) {
            throw new \InvalidArgumentException('Only GitHub provider is supported!', 1);
        }

        $baseUrl = "https://api.github.com/repos/{$name}/{$repository}";
        $headers = [
            'Authorization' => "Bearer {$this->token}",
            'Accept' => 'application/vnd.github.v3+json',
        ];

        $repoData = HttpService::makeRequest('get', $baseUrl, $headers, 'repository');

        $branchesData = HttpService::makeRequest('get', "{$baseUrl}/branches", $headers, 'branches');
        $branches = array_map(fn ($branch) => $branch['name'], $branchesData);

        return new ApiResponse(true, [
            'branches' => $branches,
            'avatar_url' => $repoData['owner']['avatar_url'] ?? '',
            'url' => $repoData['clone_url'] ?? '',
        ]);
    }

    public function getLatestCommitMessage(string $name, string $repository, string $provider): array
    {
        if (parent::getRemoteType($provider) != RemoteType::GitHub) {
            throw new \InvalidArgumentException('Only GitHub provider is supported!', 1);
        }

        $baseUrl = "https://api.github.com/repos/{$name}/{$repository}";
        $headers = [
            'Authorization' => "Bearer {$this->token}",
            'Accept' => 'application/vnd.github.v3+json',
        ];
        $commitHeaders = $headers + ['X-GitHub-Api-Version' => '2022-11-28'];
        $commitsData = HttpService::makeRequest('get', "{$baseUrl}/commits", $commitHeaders, 'commits');

        return [
            'message' => $commitsData[0]['commit']['message'],
            'avatar_url' => $commitsData[0]['author']['avatar_url'],
        ];
    }

    public function clone(string $name, string $repository, string $provider, string $local_dir)
    {
        if (parent::getRemoteType($provider) != RemoteType::GitHub) {
            throw new \InvalidArgumentException('Only GitHab provider is supported!', 1);
        }

        $cmd = "git clone https://{$this->token}@github.com/$name/$repository $local_dir";

        $result = Cmd::execute($cmd);

        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput);
        }
    }
}

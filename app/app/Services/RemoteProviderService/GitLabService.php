<?php

namespace App\Services\RemoteProviderService;

use App\Services\CmdService;
use App\Services\HttpService;
use App\DTOs\ApiResponse;
use App\Facades\AppConfig;
use App\Enums\RemoteType;

class GitLabService extends RemoteProviderService {
    public function __construct()
    {
        $token = AppConfig::get('gitlab_token');

        if (empty($token) || !is_string($token)) 
        {
            throw new \RuntimeException("GitLab token required. Check Settings > Configuration.", 1);
        }

        parent::__construct($token);
    }

    public function search(string $name, string $repository, string $provider): ApiResponse
    {
        if (parent::getRemoteType($provider) != RemoteType::GitLab)
        {
            throw new \InvalidArgumentException("Only GitLab provider is supported!", 1);
        }

        $baseUrl = "https://gitlab.com/api/v4/projects/{$name}%2F{$repository}";
        $headers = [
            'Authorization' => "Bearer {$this->token}",
            'Accept'        => 'application/json',
        ];

        $repoData = HttpService::makeRequest('get', $baseUrl, $headers, 'repository');

        $branchesData = HttpService::makeRequest('get', "{$baseUrl}/repository/branches", $headers, 'branches');
        $branches = array_map(fn($branch) => $branch['name'], $branchesData);

        return new ApiResponse(true, [
            'branches' => $branches,
            'avatar_url' => $repoData['namespace']['avatar_url'] ?? '',
            'url' => $repoData['web_url'] ?? '',
        ]);
    }

    public function getLatestCommitMessage(string $name, string $repository, string $provider): array
    {
        if (parent::getRemoteType($provider) != RemoteType::GitLab)
        {
            throw new \InvalidArgumentException("Only GitLab provider is supported!", 1);
        }

        $baseUrl = "https://gitlab.com/api/v4/projects/$name%2F$repository";
        $headers = [
            'Authorization' => "Bearer {$this->token}",
            'Accept'        => 'application/json',
        ];

        $repoData = HttpService::makeRequest('get', $baseUrl, $headers, 'repository');

        $commitHeaders = $headers;
        $commitsData = HttpService::makeRequest('get', "$baseUrl/repository/commits", $commitHeaders, 'commits');

        return [
            "message" => $commitsData[0]['message'], 
            "avatar_url" => $repoData['namespace']['avatar_url'] ?? '',
        ];
    }

    public function clone(string $name, string $repository, string $provider, string $local_dir) 
    {
        if (parent::getRemoteType($provider) != RemoteType::GitLab)
        {
            throw new \InvalidArgumentException("Only GitLab provider is supported!", 1);
        }

        $cmd = "git clone https://oauth2:{$this->token}@gitlab.com/$name/$repository $local_dir";

        $result = CmdService::execute($cmd);        
        
        if ($result->exitCode !== 0) 
        {
            throw new \RuntimeException($result->errorOutput);
        } 
    }
}

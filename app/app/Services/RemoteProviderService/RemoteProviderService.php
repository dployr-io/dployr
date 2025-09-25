<?php

namespace App\Services\RemoteProviderService;

use App\DTOs\ApiResponse;
use App\Enums\RemoteType;

abstract class RemoteProviderService {
    protected string $token;

    public function __construct(string $token) {
        $this->token = $token;
    }

    /**
     * Search for a remote repository, e.g. https://github.com/owner/repo.git
     * @param string $name The repository name (e.g. "repo" in the above example)
     * @param string $repository The owner or organization name (e.g. "owner" in the above example)
     * @param string $token Personal access token (e.g. for GitHub, a PAT with appropriate permissions)
     * @return ApiResponse
     */
    abstract protected function search(string $name, string $repository, string $provider): ApiResponse;

    abstract protected function getLatestCommitMessage(string $name, string $repository, string $provider): array;

    abstract protected function clone(string $name, string $repository, string $provider, string $local_dir);

    public static function getRemoteType(string $url): RemoteType
    {
        return match($url) {
            'github.com'   => RemoteType::GitHub,
            'gitlab.com'   => RemoteType::GitLab,
            'bitbucket.org'=> RemoteType::BitBucket,
            default        => throw new \InvalidArgumentException("Unsupported remote host: $url", 1),
        };
    }
}
<?php

namespace App\Services\RemoteProviderServices;

use App\DTOs\RemoteRepo;

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
     * @return RemoteRepo
     */
    abstract public function search(string $name, string $repository): RemoteRepo;
}
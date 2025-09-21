<?php

namespace App\DTOs;

use App\Enums\RemoteType;

readonly class RemoteRepo
{
    public function __construct(
        /** Repository name in the format "foo/bar" */
        public string $name,

        public RemoteType $remote_type,

        /** List of all branches, e.g. ["main", "staging", "master"] */
        public array $branches,

        /** Direct URL to the repository */
        public string $url,

        /** URL to the repository avatar image */
        public string $avatar_url,

        /** Last commit message */
        public ?string $commit_message = null,
    ) {
    }
}
<?php

namespace App\DTOs;

use App\Models\Remote;

readonly class Spec
{
    public function __construct(
        public string $id,
        public string $serviceName,
        public ?Remote $remote,
        public ?string $commitHash,
        public ?Remote $ciRemote,
        public string $source,
        public string $runtime,
        public ?int $port,
        public ?string $image,
        public array $envVars,
        public array $secrets,
        public ?string $staticDir,
        public ?string $workingDir,
        public ?string $runCmd,
    ) {}
}

<?php

namespace App\Services\Runtime;

use App\Constants\Runtimes;
use App\Contracts\Services\ListRuntimeVersionsInterface;
use App\Contracts\Services\SetupRuntimeInterface;
use App\Services\Cmd;

class GoService implements ListRuntimeVersionsInterface, SetupRuntimeInterface
{
    public function __construct(
        public string $version = 'latest',
    ) {
        $this->version = $version;
    }

    public function setup(string $path): void
    {
        $result = Cmd::execute("bash -lc 'asdf install golang {$this->version}'", [
            'working_dir' => $path,
            'timeout' => 900,
        ]);

        if (! $result->successful) {
            throw new \RuntimeException('Failed to install '.Runtimes::GO." {$this->version}. {$result->output}");
        }

        $result = Cmd::execute("bash -lc 'asdf set golang {$this->version}'", ['working_dir' => $path]);

        if (! $result->successful) {
            throw new \RuntimeException('Failed to set '.Runtimes::GO." {$this->version}. {$result->output}");
        }
    }

    public function list(): array
    {
        $result = Cmd::execute("bash -lc 'asdf list all golang'");

        if (! $result->successful) {
            throw new \RuntimeException("Error Processing Request {$result->errorOutput}", 1);
        }

        $values = array_values(array_filter(array_map('trim', explode("\n", $result->output))));

        return collect($values)
            ->unique()
            ->sort(fn ($a, $b) => version_compare($b, $a)) // descending
            ->values()
            ->toArray();
    }
}

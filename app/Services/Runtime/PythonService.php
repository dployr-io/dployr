<?php

namespace App\Services\Runtime;

use App\Constants\Runtimes;
use App\Contracts\Services\ListRuntimeVersionsInterface;
use App\Contracts\Services\SetupRuntimeInterface;
use App\Services\Cmd;

class PythonService implements ListRuntimeVersionsInterface, SetupRuntimeInterface
{
    public function __construct(
        public string $version = 'latest',
    ) {
        $this->version = $version;
    }

    public function setup(string $path): void
    {
        $result = Cmd::execute("bash -lc 'asdf install python {$this->version}'", [
            'working_dir' => $path,
            'timeout' => 900,
        ]);

        if (! $result->successful) {
            throw new \RuntimeException('Failed to install '.Runtimes::PYTHON." {$this->version}. {$result->output}");
        }

        $result = Cmd::execute("bash -lc 'asdf set python {$this->version}'", ['working_dir' => $path]);

        if (! $result->successful) {
            throw new \RuntimeException('Failed to set '.Runtimes::PYTHON." {$this->version}. {$result->output}");
        }
    }

    public function list(): array
    {
        $result = Cmd::execute("bash -lc 'asdf list all python'");

        if (! $result->successful) {
            throw new \RuntimeException("Error Processing Request {$result->errorOutput}", 1);
        }

        $lines = array_map(
            fn ($s) => trim(preg_replace('/[\x00-\x1F\x7F]+/', '', $s)),
            explode("\n", $result->output)
        );
        $values = array_values(array_filter($lines, fn ($s) => $s !== ''));

        $keepRegex = '/^(?=\d)\d+\.\d+(?:\.\d+)?(?:[a-z0-9.\-]*)?$/i';

        return collect($values)
            ->filter(fn ($v) => (bool) preg_match($keepRegex, $v))
            ->unique()
            ->sort(fn ($a, $b) => version_compare($b, $a)) // descending
            ->values()
            ->toArray();
    }
}

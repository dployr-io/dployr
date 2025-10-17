<?php

namespace App\Services\Runtime;

use App\Constants\Runtimes;
use App\Contracts\Services\ListRuntimeVersionsInterface;
use App\Contracts\Services\SetupRuntimeInterface;

final class RuntimeService implements ListRuntimeVersionsInterface, SetupRuntimeInterface
{
    public function __construct(
        public string $runtime = Runtimes::NODE_JS,
        public string $version = 'latest',
    ) {}

    public function setup(string $path): void
    {
        switch ($this->runtime) {
            case Runtimes::NODE_JS:
                $nodeJs = new NodeJsService($this->version);
                $nodeJs->setup($path);
                break;
            case Runtimes::PYTHON:
                $python = new PythonService($this->version);
                $python->setup($path);
                break;
            case Runtimes::STATIC:
                break;
            default:
                throw new \RuntimeException("Invalid runtime: {$this->runtime}. Choose one of ".implode(', ', Runtimes::RUNTIMES));
        }
    }

    public function list(): array
    {
        switch ($this->runtime) {
            case Runtimes::NODE_JS:
                $nodeJs = new NodeJsService;
                $versions = $nodeJs->list();
                break;
            case Runtimes::PYTHON:
                $python = new PythonService;
                $versions = $python->list();
                break;
            case Runtimes::STATIC:
                break;
            default:
                throw new \RuntimeException("Invalid runtime: {$this->runtime}. Choose one of ".implode(', ', Runtimes::RUNTIMES));
        }

        return $versions;
    }
}

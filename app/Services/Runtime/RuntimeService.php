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
            case Runtimes::RUBY:
                $ruby = new RubyService($this->version);
                $ruby->setup($path);
                break;
            case Runtimes::JAVA:
                $java = new JavaService($this->version);
                $java->setup($path);
                break;
            case Runtimes::DOTNET:
                $dotnet = new DotNetService($this->version);
                $dotnet->setup($path);
                break;
            case Runtimes::GO:
                $go = new GoService($this->version);
                $go->setup($path);
                break;
            case Runtimes::PHP:
                $php = new PhpService($this->version);
                $php->setup($path);
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
            case Runtimes::RUBY:
                $ruby = new RubyService;
                $versions = $ruby->list();
                break;
            case Runtimes::JAVA:
                $java = new JavaService();
                $versions = $java->list();
                break;
            case Runtimes::DOTNET:
                $dotnet = new DotNetService();
                $versions = $dotnet->list();
                break;
            case Runtimes::GO:
                $go = new GoService();
                $versions = $go->list();
                break;
            case Runtimes::PHP:
                $php = new PhpService();
                $versions = $php->list();
                break;
            case Runtimes::STATIC:
                break;
            default:
                throw new \RuntimeException("Invalid runtime: {$this->runtime}. Choose one of ".implode(', ', Runtimes::RUNTIMES));
        }

        return $versions;
    }
}

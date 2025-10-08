<?php

namespace App\DTOs;

readonly class CmdResult
{
    public function __construct(
        public string $command,
        public ?int $exitCode,
        public string $output,
        public string $errorOutput,
        public bool $successful,
        public bool $isAsync = false,
    ) {}

    public function failed(): bool
    {
        return !$this->successful;
    }

    public function hasOutput(): bool
    {
        return !empty($this->output);
    }

    public function hasError(): bool
    {
        return !empty($this->errorOutput);
    }
}
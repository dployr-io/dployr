<?php

namespace App\Contracts\Services;

interface SetupRuntimeInterface
{
    public function setup(string $path): void;
}
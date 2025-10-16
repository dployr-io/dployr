<?php

namespace App\Contracts\Services;

interface ListRuntimeVersionsInterface
{
    public function list(): array;
}
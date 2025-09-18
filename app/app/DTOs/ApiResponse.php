<?php

namespace App\DTOs;

readonly class ApiResponse
{
    public function __construct(
        public bool $success,
        public array $data = [],
        public ?string $error = null
    ) {}
}
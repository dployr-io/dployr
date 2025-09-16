<?php

namespace App\DTO;

class ApiResponse
{
    public function __construct(
        public bool $success,
        public array $data = [],
        public ?string $error = null
    ) {}
}
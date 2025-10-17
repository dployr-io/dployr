<?php

namespace App\Contracts\Blueprints;

interface BlueprintServiceInterface
{
    public function getConfig(): array;

    public function getMetadata(): array;

    public function validate(): void;

    public function process(): void;
}

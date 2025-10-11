<?php

namespace App\Services\Blueprint;

use App\Services\SchemaValidatorService;

class BlueprintValidatorService extends SchemaValidatorService
{
    private const SCHEMA_URL = 'https://dployr.dev/schema/1.0.0/service.json';

    private const CACHE_KEY = 'service_schema_v1.0.0';

    private const CACHE_TTL = 86400; // 24 hours

    private function __construct(
        private string $schema_url = self::SCHEMA_URL,
        private string $cache_key = self::CACHE_KEY,
        private int $cache_ttl = self::CACHE_TTL,
    ) {
        parent::__construct($schema_url, $cache_key, $cache_ttl);
    }

    public function validate(array $config): array
    {
        return parent::validate($config);
    }

    public function formatErrors(array $errors): string
    {
        return parent::formatErrors($errors);
    }

    public function clearCache(): void
    {
        return parent::clearCache();
    }
}

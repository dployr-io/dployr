<?php

namespace App\Services\Blueprint;

use App\Services\SchemaValidatorService;

class BlueprintValidatorService extends SchemaValidatorService
{
    protected const SCHEMA_URL = 'https://dployr.dev/schema/1.0.0/service.json';

    protected const CACHE_KEY = 'service_schema_v1.0.0';

    protected const CACHE_TTL = 86400; // 24 hours

    public function __construct(
        protected string $schema_url = self::SCHEMA_URL,
        protected string $cache_key = self::CACHE_KEY,
        protected int $cache_ttl = self::CACHE_TTL,
    ) {
        parent::__construct($schema_url, $cache_key, $cache_ttl);
    }

    public function validate(mixed $config)
    {
        parent::validate($config);
    }

    public function formatErrors(array $errors): string
    {
        return parent::formatErrors($errors);
    }

    public function clearCache(): void
    {
        parent::clearCache();
    }
}

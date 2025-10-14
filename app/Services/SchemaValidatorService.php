<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use Swaggest\JsonSchema\Schema;

abstract class SchemaValidatorService
{
    public function __construct(
        protected string $schema_url,
        protected string $cache_key,
        protected int $cache_ttl
    ) {}

    /**
     * Validate the blueprint config against the JSON schema
     *
     * @param  mixed  $config  The config to validate
     */
    public function validate(mixed $config)
    {
        $schema = $this->getSchema();
        
        try {
            $schema = Schema::import($schema)->in($config);  
        } catch (\Exception $e) {
            echo "JSON validation error: " . $e->getMessage();
        }
    }

    /**
     * Get validation errors as a formatted string
     *
     * @param  array  $errors  The errors array from validate()
     */
    public function formatErrors(array $errors): string
    {
        $formatted = [];
        foreach ($errors as $error) {
            $property = $error['property'] ?: 'root';
            $formatted[] = "- {$property}: {$error['message']}";
        }

        return implode("\n", $formatted);
    }

    /**
     * Fetch and cache the JSON schema
     *
     * @throws \RuntimeException
     */
    private function getSchema(): string
    {
        $schema = Cache::remember($this->cache_key, $this->cache_ttl, function () {
            try {
                $response = HttpService::makeRequest('get', $this->schema_url);

                if (is_bool($response)) {
                    throw new \RuntimeException('Invalid schema: Received boolean value');
                }
                return $response;
            } catch (\Exception $e) {
                throw new \RuntimeException(
                    'Unable to fetch schema from '.$this->schema_url.': '.$e->getMessage()
                );
            }
        });
        
        return json_encode($schema);
    }

    /**
     * Clear the cached schema (useful for testing or when schema updates)
     */
    public function clearCache(): void
    {
        Cache::forget($this->schema_url);
    }
}

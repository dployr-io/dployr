<?php

namespace App\Services;

use Illuminate\Support\Facades\Cache;
use JsonSchema\Constraints\Constraint;
use JsonSchema\Validator;

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
     * @param  array  $config  The config array to validate
     * @return array Returns ['valid' => bool, 'errors' => array]
     */
    public function validate(array $config): array
    {
        $schema = $this->getSchema();
        $configObject = json_decode(json_encode($config));
        $validator = new Validator;
        $validator->validate(
            $configObject,
            $schema,
            Constraint::CHECK_MODE_APPLY_DEFAULTS
        );

        if ($validator->isValid()) {
            return [
                'valid' => true,
                'errors' => [],
            ];
        }
        $errors = [];
        foreach ($validator->getErrors() as $error) {
            $errors[] = [
                'property' => $error['property'],
                'message' => $error['message'],
                'constraint' => $error['constraint'] ?? null,
            ];
        }

        return [
            'valid' => false,
            'errors' => $errors,
        ];
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
    private function getSchema(): object
    {
        $schema = Cache::remember($this->cache_key, $this->cache_ttl, function () {
            try {
                return HttpService::makeRequest('get', $this->schema_url);
            } catch (\Exception $e) {
                throw new \RuntimeException(
                    'Unable to fetch schema from '.$this->schema_url.': '.$e->getMessage()
                );
            }
        });

        return json_decode(json_encode($schema));
    }

    /**
     * Clear the cached schema (useful for testing or when schema updates)
     */
    public function clearCache(): void
    {
        Cache::forget($this->schema_url);
    }
}

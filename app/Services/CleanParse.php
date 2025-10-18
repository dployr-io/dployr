<?php

namespace App\Services;

class CleanParse
{
    /**
     * Strips out null values from an object
     */
    public static function withoutNulls(array $data): array
    {
        return array_filter($data, fn ($v) => ! is_null($v));
    }
}

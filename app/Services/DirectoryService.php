<?php

namespace App\Services;

class DirectoryService
{
    public static function setupFolder(string $path): bool
    {
        $cmd = "mkdir -p {$path}";

        $result = Cmd::execute($cmd);

        return $result === 0;
    }
}

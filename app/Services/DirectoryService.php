<?php

namespace App\Services;

class DirectoryService
{
    public static function setupFolder(string $path) : bool 
    {
        $cmd = "mkdir -p {$path}";  

        $result = CmdService::execute($cmd);

        return $result === 0;
    }
}

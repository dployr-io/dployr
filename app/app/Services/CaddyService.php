<?php

namespace App\Services;

class CaddyService
{
    public static function getStatus(): string
    {
        $result = CmdService::execute("cat /etc/caddy/Caddyfile");

        return $result->output;
    }
}
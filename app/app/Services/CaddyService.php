<?php

namespace App\Services;

class CaddyService
{
    public static function status(): bool
    {
        $result = CmdService::execute("systemctl is-active caddy");
        return $result->output === 'active';
    }

    public static function config(): string
    {
        $path = '/etc/caddy/Caddyfile';
        $result = @file_get_contents( $path);

        if (!$result) {
            throw new \RuntimeException("Caddyfile not found or inaccessible in $path", 1);
        }
        return trim($result);
    }

    public static function append(string $block): bool
    {
        $caddyfile = self::config();
        $newContent = $caddyfile . "\n\n" . trim(str_replace(["\r\n", "\r"], "\n", $block));
        $tmp = tempnam(sys_get_temp_dir(), 'caddy');
        file_put_contents($tmp, $newContent);
        CmdService::execute("chown caddy:caddy $tmp");
        CmdService::execute("chmod 644 $tmp");
        $result = rename($tmp, '/etc/caddy/Caddyfile');
        
        if (!$result) {
            unlink($tmp);
            throw new \RuntimeException("Failed to update Caddyfile", 1);
        }
        return $result;
    }

    public static function checkPort(int $port): bool
    {
        $caddyfile = self::config();
        $pattern = "/^\\s*:$port\\b/m";
        return preg_match($pattern, $caddyfile) === 1;
    }

    public static function restart()
    {
        $result = CmdService::execute("systemctl restart caddy");
        
        if (!$result->successful) {
            throw new \RuntimeException("Failed to restart Caddy service", 1);
        }
    }
}
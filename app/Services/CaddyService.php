<?php

namespace App\Services;

class CaddyService
{
    public static function status(): bool
    {
        $result = CmdService::execute('systemctl is-active caddy');
        return $result->output === 'active';
    }

    public static function config(): string
    {
        $path = '/etc/caddy/Caddyfile';
        $result = @file_get_contents($path);

        if (! $result) {
            throw new \RuntimeException("Caddyfile not found or inaccessible in $path", 1);
        }
        return trim($result);
    }

    public static function newConfig(string $serviceName, string $block)
    {
        $baseConfig = '/etc/caddy/Caddyfile';
        $sitesDir = '/etc/caddy/sites-enabled';
        $filePath = "$sitesDir/{$serviceName}.conf";
        
        if (! is_dir($sitesDir)) {
            mkdir($sitesDir, 0750, true);
        }

        $block = trim(str_replace(["\r\n", "\r"], "\n", $block));
        $tmp = tempnam(sys_get_temp_dir() . "/dployr", 'caddy');
        file_put_contents($tmp, $block);
        $result = CmdService::execute("chown caddy:caddy $tmp");

        if (! $result->successful) {
            unlink($tmp);
            throw new \RuntimeException("Failed to modify ownership for temporary config file {$tmp}, while deploying service {$serviceName}", 1);
        }
        $result = CmdService::execute("chmod 644 $tmp");

        if (! $result->successful) {
            unlink($tmp);
            throw new \RuntimeException("Failed to modify permissions for temporary config file {$tmp}, while deploying service {$serviceName}", 1);
        }
        $result = rename($tmp, $filePath);

        if (! $result) {
            unlink($tmp);
            throw new \RuntimeException("Failed to create site config for {$serviceName}", 1);
        }
        $result = CmdService::execute("caddy validate --config {$baseConfig}");

        if (! $result->successful) {
            unlink($tmp);
            \Log::warning('Rolling back config. No changes was made.');
            $_result = rename($filePath, unlink($tmp));

            if (! $_result) {
                throw new \RuntimeException('Failed to rollback to a known working configration. Access may break!');
            }
            throw new \RuntimeException("Failed to update caddy config for {$serviceName}. {$result->errorOutput}", 1);
        }
    }

    public static function checkPort(int $port): bool
    {
        $caddyfile = self::config();
        $pattern = "/^\\s*:$port\\b/m";
        return preg_match($pattern, $caddyfile) === 1;
    }

    public static function restart()
    {
        $result = CmdService::execute('systemctl restart caddy');
        if (! $result->successful) {
            throw new \RuntimeException('Failed to restart Caddy service', 1);
        }
    }
}

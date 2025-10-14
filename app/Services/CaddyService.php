<?php

namespace App\Services;

class CaddyService
{
    public function status(): bool
    {
        $result = CmdService::execute('systemctl is-active caddy');

        return $result->output === 'active';
    }

    public function newConfig(string $serviceName, string $block)
    {
        $baseConfig = '/etc/caddy/Caddyfile';
        $sitesDir = '/etc/caddy/sites-enabled';
        $filePath = "$sitesDir/{$serviceName}.conf";
        $tmpDir = '/home/dployr/tmp';

        if (! is_dir($sitesDir)) {
            mkdir($sitesDir, 0750, true);
        }

        if (! is_dir($tmpDir)) {
            $oldUmask = umask(0);
            mkdir($tmpDir, 01777, true);
            umask($oldUmask);
        }

        $block = trim(str_replace(["\r\n", "\r"], "\n", $block));
        $tmp = tempnam($tmpDir, 'caddy');
        file_put_contents($tmp, $block);
        $result = CmdService::execute("chown dployr:caddy $tmp");

        if (! $result->successful) {
            unlink($tmp);
            throw new \RuntimeException("Failed to modify ownership for temporary config file {$tmp}, while deploying service {$serviceName} {$result->errorOutput}", 1);
        }
        $result = CmdService::execute("chmod 644 $tmp");

        if (! $result->successful) {
            unlink($tmp);
            throw new \RuntimeException("Failed to modify permissions for temporary config file {$tmp}, while deploying service {$serviceName} {$result->errorOutput}", 1);
        }
        $result = rename($tmp, $filePath);

        if (! $result) {
            unlink($tmp);
            throw new \RuntimeException("Failed to create site config for {$serviceName}", 1);
        }
        $result = CmdService::execute("caddy validate --config {$baseConfig} --adapter caddyfile");

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

    private function processedConfig(): string
    {
        $result = CmdService::execute('curl localhost:2019/config/ | jq');

        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput, 1);
        }

        return trim($result->output);
    }

    public function checkPort(int $port): bool
    {
        $caddyfile = self::processedConfig();
        $pattern = "/:$port\b/";

        return preg_match($pattern, $caddyfile) === 1;
    }

    public function restart()
    {
        $result = CmdService::execute('sudo systemctl restart caddy');
        if (! $result->successful) {
            throw new \RuntimeException('Failed to restart Caddy service', 1);
        }
    }
}

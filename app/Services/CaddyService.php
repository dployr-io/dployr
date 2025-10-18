<?php

namespace App\Services;

use App\Constants\Runtimes;

class CaddyService
{
    public function status(): bool
    {
        $result = Cmd::execute('systemctl is-active caddy');

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
        $result = Cmd::execute("chown dployr:caddy $tmp");

        if (! $result->successful) {
            unlink($tmp);
            throw new \RuntimeException("Failed to modify ownership for temporary config file {$tmp}, while deploying service {$serviceName} {$result->errorOutput}", 1);
        }
        $result = Cmd::execute("chmod 644 $tmp");

        if (! $result->successful) {
            unlink($tmp);
            throw new \RuntimeException("Failed to modify permissions for temporary config file {$tmp}, while deploying service {$serviceName} {$result->errorOutput}", 1);
        }
        $result = rename($tmp, $filePath);

        if (! $result) {
            unlink($tmp);
            throw new \RuntimeException("Failed to create site config for {$serviceName}", 1);
        }
        $result = Cmd::execute("caddy validate --config {$baseConfig} --adapter caddyfile");

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

    public function checkPort(int $port): bool
    {
        $caddyfile = self::processedConfig();
        $pattern = "/:$port\b/";

        return preg_match($pattern, $caddyfile) === 1;
    }

    public function restart()
    {
        $result = Cmd::execute('sudo systemctl restart caddy');
        if (! $result->successful) {
            throw new \RuntimeException('Failed to restart Caddy service', 1);
        }
    }

    private function processedConfig(): string
    {
        $result = Cmd::execute('curl localhost:2019/config/ | jq');

        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput, 1);
        }

        return trim($result->output);
    }

    /**
     * Create a new Caddy configuration block for a service.
     *
     * @param string|null $staticPath Path to static files (if any)
     * @param int $port Public port exposed by Caddy
     * @param int|null $servicePort Internal port where app runs (for dynamic runtimes)
     * @param string $runtime Service runtime type
     * @return string Generated Caddy config block
     */
    public function newBlock(
        ?string $staticPath,
        int $port,
        ?int $servicePort,
        string $runtime
    ): string {
        $hasStatic = !empty($staticPath);
        $isDynamic = in_array($runtime, [
            Runtimes::GO,
            Runtimes::PHP,
            Runtimes::PYTHON,
            Runtimes::NODE_JS,
            Runtimes::RUBY,
            Runtimes::DOTNET,
            Runtimes::JAVA,
        ]);

        if ($runtime === Runtimes::STATIC) {
            return <<<EOF
:$port {
    root * {$staticPath}
    encode zstd gzip

    file_server
    header Cache-Control "public, max-age=31536000, immutable"
}
EOF;
        }

        $staticBlock = "";
        $serviceBlock = "";

        if ($hasStatic) {
            $staticBlock = <<<EOF
    root * {$staticPath}
    encode zstd gzip

    handle_path /assets/* {
        file_server
        header Cache-Control "public, max-age=31536000, immutable"
    }
EOF;
        }

        $reverseProxy = $runtime === Runtimes::PHP
            ? "php_fastcgi localhost:{$servicePort}"
            : "reverse_proxy localhost:{$servicePort}";
        
        if ($isDynamic) {
            $serviceBlock = <<<EOF
    handle {
        {$reverseProxy}
    }
EOF;
        }

        return <<<EOF
:$port {
    {$staticBlock} 
    {$serviceBlock}
}
EOF;
    }
}

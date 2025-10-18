<?php

namespace App\Services\Secrets;

use App\Services\Cmd;
use Illuminate\Support\Facades\File;

final class SecretsManagerService
{
    private string $envFile;
    private string $secretsFile;

    public function __construct() {}

    /** Initializes a new env file and secrets file */
    public function init(string $workingDir, string $name): void
    {
        $tmpDir = "/home/dployr/tmp/{$name}";
        $tmpEnv = "{$tmpDir}/.env";
        $tmpSecrets = "{$tmpDir}/.env.secrets";

        $commands = [];

        if (File::exists($tmpEnv) && File::exists($tmpSecrets)) {
            $commands = [
                "mkdir -p $(dirname {$workingDir})",
                "cp {$tmpEnv} {$workingDir}/.env",
                "cp {$tmpSecrets} {$workingDir}/.env.secrets",
                "chmod 640 {$workingDir}/.env",
                "chmod 600 {$workingDir}/.env.secrets",
                "rm -rf {$tmpDir}",
            ];
        } else {
            $commands = [
                ": > {$workingDir}/.env",
                ": > {$workingDir}/.env.secrets",
                "chmod 640 {$workingDir}/.env",
                "chmod 600 {$workingDir}/.env.secrets",
                "rm -rf {$tmpDir}",
            ];
        }

        $bash = implode(' && ', $commands);

        $result = Cmd::execute("bash -lc \"$bash\"");

        if (! $result->successful) {
            throw new \RuntimeException("Failed to initialize env files: {$result->errorOutput}");
        }
    }
    
    /** Create or update normal env variables */
    public function setEnv(array $variables): void
    {
        $this->writeEnvFile($this->envFile, $variables);

        $result = Cmd::execute("chmod 640 {$this->envFile}");
        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput);
        }
    }

    /** Create or update secret variables */
    public function setSecrets(array $secrets): void
    {
        $this->writeEnvFile($this->secretsFile, $secrets);
        
        $result = Cmd::execute("chmod 600 {$this->envFile}");
        if (! $result->successful) {
            throw new \RuntimeException($result->errorOutput);
        }
    }

    /** Read all normal env variables */
    public function getEnv(): array
    {
        return $this->parseEnvFile($this->envFile);
    }

    /** Read secret keys only (hide values) */
    public function getSecretKeys(): array
    {
        $secrets = $this->parseEnvFile($this->secretsFile);
        return array_keys($secrets);
    }

    /** Stores a temporary .env file for a service */
    public function tmp(mixed $envVars, mixed $secrets, string $name): void
    {
        $tmpDir = "/home/dployr/tmp/{$name}";
        $envFile = "{$tmpDir}/.env";
        $secretsFile = "{$tmpDir}/.env.secrets";

        $commands = [
            "mkdir -p {$tmpDir}",
            "rm -f {$envFile} {$secretsFile}",
            "touch {$envFile} {$secretsFile}",
            "chmod 640 {$envFile}",
            "chmod 600 {$secretsFile}",
        ];

        foreach ($envVars as $key => $value) {
            $commands[] = "echo {$key}=" . escapeshellarg($value) . " >> {$envFile}";
        }

        foreach ($secrets as $key => $value) {
            $commands[] = "echo {$key}=" . escapeshellarg($value) . " >> {$secretsFile}";
        }

        $bash = implode(' && ', $commands);

        Cmd::execute("bash -lc \"$bash\"", ['async' => true]);
    }


    private function parseEnvFile(string $path): array
    {
        if (! File::exists($path)) {
            return [];
        }

        $lines = file($path, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);
        $vars = [];

        foreach ($lines as $line) {
            if (str_starts_with(trim($line), '#') || !str_contains($line, '=')) {
                continue;
            }

            [$key, $value] = explode('=', $line, 2);
            $vars[trim($key)] = trim($value, "\"' ");
        }

        return $vars;
    }

    private function writeEnvFile(string $path, array $vars): void
    {
        $content = '';
        foreach ($vars as $key => $value) {
            $content .= sprintf("%s=\"%s\"\n", $key, addslashes($value));
        }

        $result = File::put($path, $content);

        if (! $result) {
            throw new \RuntimeException("Failed to write to env file {$path}");
        }
    }
}
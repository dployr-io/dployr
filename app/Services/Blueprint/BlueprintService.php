<?php

namespace App\Services\Blueprint;

use App\Constants\Runtimes;
use App\Contracts\Blueprints\BlueprintServiceInterface;
use App\Models\Blueprint;
use App\Models\Remote;
use App\Models\Service;
use App\Services\CaddyService;
use App\Services\CleanParse;
use App\Services\Cmd;
use App\Services\DirectoryService;
use App\Services\GitRepoService;
use App\Services\Runtime\RuntimeService;
use App\Services\Secrets\SecretsManagerService;
use App\Services\SystemdService;
use Illuminate\Support\Facades\Log;

class BlueprintService implements BlueprintServiceInterface
{
    protected const BASE_PATH = '/home/dployr/services';

    public function __construct(
        private Blueprint $blueprint,
    ) {
        $this->blueprint = $blueprint;
    }

    public function getConfig(): array
    {
        $config = $this->blueprint->config;

        return is_array($config) ? $config : json_decode($config, true) ?? [];
    }

    public function getMetadata(): array
    {
        $metadata = $this->blueprint->metadata;

        return is_array($metadata) ? $metadata : json_decode($metadata, true) ?? [];
    }

    public function getRuntime(mixed $runtime): array
    {
        return is_array($runtime) ? $runtime : json_decode($runtime, true) ?? [];
    }

    public function validate(): void
    {
        $validator = new BlueprintValidatorService;

        $validator->validate($this->getConfig());
    }

    public function process(): void
    {
        $currentStatus = $this->blueprint->fresh()->status;

        if ($currentStatus !== 'pending') {
            Log::debug("Blueprint {$this->blueprint->id} status is {$currentStatus}, skipping");

            return;
        }

        $updated = $this->blueprint->where('status', 'pending')->update(['status' => 'in_progress']);

        if (! $updated) {
            Log::debug("Blueprint {$this->blueprint->id} status changed by another process, skipping");

            return;
        }

        try {
            // TODO: Ensure validation is done on the blueprint to
            // be sure that the runtime selected, matches the right resource
            $this->validate();

            $config = $this->getConfig();
            $metadata = $this->getMetadata();
            $runtime = $this->getRuntime($config['runtime']);
            $remoteId = $config['remote'];
            $port = $config['port'];
            $runCmd = $config['run_cmd'] ?? null;
            $buildCmd = $config['build_cmd'] ?? null;
            $name = $config['name'];
            $path = rtrim(self::BASE_PATH, '/').'/'.$config['name'].'/';
            $workingDir = ltrim($config['working_dir'] ?? '', '/');
            $staticDir = ltrim($config['static_dir'] ?? '', '/');
            $staticPath = $path.$workingDir.'/'.$staticDir;
            $servicePort = $this->getServicePort($name, $runtime['type']);

            Log::info('Blueprint config validation successful');

            DirectoryService::setupFolder($path);
            $remote = Remote::findOrFail($remoteId);

            $remoteService = new GitRepoService;
            $remoteService->cloneRepo($remote->name, $remote->repository, $remote->provider, $path);

            $appRuntime = new RuntimeService($runtime['type'], $runtime['version']);
            $appRuntime->setup($path.$workingDir);

            if ($buildCmd !== null) {
                $cmd = Cmd::execute("bash -lc '{$buildCmd}'", ['working_dir' => $path.$workingDir]);
                $result = $cmd->successful;

                if (! $result) {
                    throw new \RuntimeException("Build command failed: $buildCmd");
                }
            }

            $secretsManager = new SecretsManagerService;
            $secretsManager->init($path.$workingDir, $name);

            $systemd = new SystemdService;
            $systemd->newService($name, $path.$workingDir, $runCmd);

            $caddy = new CaddyService;
            $block = $caddy->newBlock($staticPath, $port, $servicePort, $runtime['type']);
            $caddy->newConfig($name, $block);
            $caddy->restart();
            $result = $this->blueprint->updateOrFail(['status' => 'completed']);

            if (! $result) {
                throw new \RuntimeException("Failed to update blueprint status for ID {$this->blueprint->id}", 1);
            }

            $service = Service::create(CleanParse::withoutNulls(
                [
                    'name' => $name,
                    'source' => $config['source'],
                    'runtime' => $runtime['type'],
                    'runtime_version' => $runtime['version'] ?? null,
                    'run_cmd' => $runCmd,
                    'build_cmd' => $buildCmd,
                    'port' => $port,
                    'working_dir' => $workingDir,
                    'static_dir' => $staticDir,
                    'image' => $config['image'] ?? null,
                    // 'env_vars' => $spec->envVars,
                    // 'secrets' => $spec->secrets,
                    'remote_id' => $remote->id,
                    'ci_remote_id' => $config['ci_remote'] ?? null,
                    'project_id' => $metadata['project_id'] ?? null,
                ]
            ));

            Log::info('Successfully created service '.$config['name'].' ID: '.$service->id);
        } catch (\RuntimeException $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            $config = $this->getConfig();
            Log::error('Runtime exception on service '.$config['name'].' '.$e->getMessage());
        } catch (\Exception $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            $config = $this->getConfig();
            $errorMessage = $e instanceof \Throwable ? $e->getMessage() : 'An unexpected error occurred.';
            Log::error('Failed to create service '.$config['name']." $errorMessage");
        }
    }

    private function getServicePort(string $name, $runtime): ?string
    {
        $tmpEnv = "/home/dployr/tmp/$name/.env";

        if ($runtime === Runtimes::STATIC || $runtime === Runtimes::K3S || $runtime === Runtimes::DOCKER) {
            return null;
        }

        if (file_exists($tmpEnv)) {
            $envContent = file_get_contents($tmpEnv);
            if (preg_match('/^PORT\s*=\s*(\d+)/m', $envContent, $matches)) {
                $servicePort = (int) $matches[1];

                Log::debug("Found port $servicePort for $name");
            }
        }

        if ($servicePort === null) {
            do {
                $servicePort = rand(7000, 7999);
            } while ($servicePort === 7879);
        }

        return $servicePort;
    }
}

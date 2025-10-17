<?php

namespace App\Services\Blueprint;

use App\Contracts\Blueprints\BlueprintServiceInterface;
use App\Models\Blueprint;
use App\Models\Remote;
use App\Models\Service;
use App\Services\CaddyService;
use App\Services\CleanParseService;
use App\Services\Cmd;
use App\Services\DirectoryService;
use App\Services\GitRepoService;
use App\Services\Runtime\RuntimeService;
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
            $path = rtrim(self::BASE_PATH, '/').'/'.$config['name'].'/';
            $workingDir = ltrim($config['working_dir'] ?? '', '/');
            $staticDir = ltrim($config['static_dir'] ?? '', '/');
            $staticPath = $path.$workingDir.'/'.$staticDir;

            Log::info('Blueprint config validation successful');

            $this->blueprint->updateOrFail(['status' => 'in_progress']);

            DirectoryService::setupFolder($path);

            $remoteId = $config['remote'];
            $remote = Remote::findOrFail($remoteId);

            $remoteService = new GitRepoService;
            $remoteService->cloneRepo($remote->name, $remote->repository, $remote->provider, $path);

            $appRuntime = new RuntimeService($runtime['type'], $runtime['version']);
            $appRuntime->setup($path);

            $port = $config['port'];
            $newBlock = <<<EOF
            :$port {
                root * $staticPath
                file_server
            }
            EOF;

            $runCmd = $config['run_cmd'] ?? null;

            if ($runCmd !== null) {
                $cmd = Cmd::execute("bash -lc '{$runCmd}'", ['working_dir' => $path.'/'.$workingDir]);
                $result = $cmd->successful;

                if (! $result) {
                    throw new \RuntimeException("Run command failed: $runCmd");
                }
            }

            $caddy = new CaddyService;
            $caddy->newConfig($config['name'], $newBlock);
            $caddy->restart();
            $result = $this->blueprint->updateOrFail(['status' => 'completed']);

            if (! $result) {
                throw new \RuntimeException("Failed to update blueprint status for ID {$this->blueprint->id}", 1);
            }

            $service = Service::create(CleanParseService::withoutNulls(
                [
                    'name' => $config['name'],
                    'source' => $config['source'],
                    'runtime' => $config['runtime'],
                    'runtime_version' => $config['runtime_version'] ?? null,
                    'run_cmd' => $config['run_cmd'] ?? null,
                    'port' => $config['port'],
                    'working_dir' => $config['working_dir'] ?? null,
                    'static_dir' => $config['static_dir'] ?? null,
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
}

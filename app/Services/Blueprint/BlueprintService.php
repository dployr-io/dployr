<?php

namespace App\Services\Blueprint;

use App\Models\Blueprint;
use App\Models\Remote;
use App\Models\Service;
use App\Services\CaddyService;
use App\Services\CleanParseService;
use App\Services\CmdService;
use App\Services\DirectoryService;
use App\Services\GitRepoService;
use Illuminate\Support\Facades\Log;

class BlueprintService
{
    protected const BASE_PATH = '/home/dployr/services';

    public function __construct(
        private Blueprint $blueprint,
    ) {
        $this->blueprint = $blueprint;
    }

    private function validateBlueprint()
    {
        $validator = new BlueprintValidatorService;

        return $validator->validate($this->getConfig());
    }

    private function getConfig(): array
    {
        $config = $this->blueprint->config;
        return is_array($config) ? $config : json_decode($config, true) ?? [];
    }

    private function getMetadata(): array
    {
        $metadata = $this->blueprint->metadata;
        return is_array($metadata) ? $metadata : json_decode($metadata, true) ?? [];
    }

    public function processBlueprint()
    {
        try {
            // TODO: Ensure validation is done on the blueprint to 
            // be sure that the runtime selected, matches the right resource
            $this->validateBlueprint();

            $config = $this->getConfig();
            $metadata = $this->getMetadata();
            $path = rtrim(self::BASE_PATH, '/').'/'.$config['name'].'/';
            $publicPath = $path.ltrim($config['working_dir'] ?? '', '/');

            Log::info("Blueprint config validation successful");
            
            $this->blueprint->updateOrFail(['status' => 'in_progress']);
            
            DirectoryService::setupFolder($path);
            
            // setup runtime 
            $remoteId = $config['remote'];
            $remote = Remote::findOrFail($remoteId);
            
            $remoteService = new GitRepoService;
            $remoteService->cloneRepo($remote->name, $remote->repository, $remote->provider, $path);

            $port = $config['port'];

            $newBlock = <<<EOF
            :$port {
                root * $publicPath
                file_server
            }
            EOF;

            $caddy = new CaddyService();
            $caddy->newConfig($config['name'], $newBlock);
            $runCmd = $config['run_cmd'] ?? null;

            if ($runCmd !== null) {
                $cmd = CmdService::execute($runCmd, ['working_directory' => $path]);
                $result = $cmd->successful;

                if (! $result) {
                    throw new \RuntimeException("Run command failed: $runCmd");
                }
            }

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
                    'output_dir' => $config['output_dir'] ?? null,
                    'image' => $config['image'] ?? null,
                    // 'env_vars' => $spec->envVars,
                    // 'secrets' => $spec->secrets,
                    'remote_id' => $remote->id,
                    'ci_remote_id' => $config['ci_remote'] ?? null,
                    'project_id' => $metadata['project_id'] ?? null,
                ]
            ));

            Log::info("Successfully created service ".$config['name']." ID: ".$service->id);
        } catch (\RuntimeException $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            $config = $this->getConfig();
            Log::error("Runtime exception on service ".$config['name']." ".$e->getMessage());
        } catch (\Exception $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            $config = $this->getConfig();
            $errorMessage = $e instanceof \Throwable ? $e->getMessage() : 'An unexpected error occurred.';
            Log::error("Failed to create service ".$config['name']." $errorMessage");
        }
    }
}

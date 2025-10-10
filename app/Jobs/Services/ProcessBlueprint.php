<?php

namespace App\Jobs\Services;

use App\Models\Blueprint;
use App\Models\Remote;
use App\Models\Service;
use App\Services\CaddyService;
use App\Services\CmdService;
use App\Services\DirectoryService;
use App\Services\GitRepoService;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Queue\Queueable;
use Illuminate\Support\Facades\Log;

class ProcessBlueprint implements ShouldQueue
{
    use Queueable;

    public function __construct(
        private Blueprint $blueprint,
    ) {}

    public function handle(): void
    {
        $id = $this->blueprint->id;
        $config = json_decode($this->blueprint->config, true);
        $serviceName = $config['name'];
        $remoteId = $config['remote'] ?? null;
        $ciRemoteId = $config['ci_remote'] ?? null;
        $runtime = $config['runtime'] ?? null;
        $port = $config['port'] ?? null;
        $image = $config['image'] ?? null;
        $spec = $config['spec'] ?? null;
        $envVars = $config['env_vars'] ?? null;
        $secrets = $config['secrets'] ?? null;
        $outputDir = $config['output_dir'] ?? null;
        $workingDir = $config['working_dir'] ?? null;
        $runCmd = $config['run_cmd'] ?? null;
        $basePath = "/home/dployr/services/$serviceName/";
        $path = $basePath.ltrim($workingDir ?? '', '/');

        try {
            $this->blueprint->updateOrFail(['status' => 'in_progress']);

            $remote = Remote::findOrFail($remoteId);
            DirectoryService::setupFolder($basePath);

            $remoteService = new GitRepoService;
            $remoteService->cloneRepo($remote->name, $remote->repository, $remote->provider, $basePath);

            $newBlock = <<<EOF
            :$port {
                root * $path/dist
                file_server
            }
            EOF;

            CaddyService::newConfig($serviceName, $newBlock);

            if ($runCmd !== null) {
                $cmd = CmdService::execute($runCmd, ['working_directory' => $path]);
                $result = $cmd->successful;

                if (! $result) {
                    throw new \RuntimeException("Run command failed: {$runCmd}");
                }
            }

            CaddyService::restart();

            $result = $this->blueprint->updateOrFail(['status' => 'completed']);

            if (! $result) {
                throw new \RuntimeException("Failed to update blueprint status for ID $id", 1);
            }

            $service = Service::create([
                'name' => $serviceName,
                'source' => $remote->repository,
                'runtime' => $runtime,
                'run_cmd' => $runCmd,
                'port' => $port,
                'working_dir' => $workingDir,
                'output_dir' => $outputDir,
                'image' => $image,
                'spec' => $spec,
                'env_vars' => $envVars,
                'secrets' => $secrets,
                'remote_id' => $remote->id,
                'ci_remote_id' => $ciRemoteId,
            ]);

            Log::info("Successfully created service $serviceName. ID: ".$service->id);
        } catch (\RuntimeException $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            Log::error("Runtime exception on service $serviceName: ".$e->getMessage());
        } catch (\Exception $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            $errorMessage = $e instanceof \Throwable ? $e->getMessage() : 'An unexpected error occurred.';
            Log::error("Failed to create service $serviceName: ".$errorMessage);
        }
    }
}

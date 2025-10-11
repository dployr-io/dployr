<?php

namespace App\Services\Blueprint;

use App\DTOs\Spec;
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

    private function parseConfig()
    {
        return json_decode($this->blueprint->config, true);
    }

    private function validateBlueprint()
    {
        $validator = new BlueprintValidatorService;

        return $validator->validate($this->parseConfig());
    }

    private function getAttributes(): Spec
    {
        $config = $this->parseConfig();

        return new Spec(
            id: $this->blueprint->id,
            serviceName: $config['name'],
            remote: $config['remote'] ?? null,
            commitHash: $config['commit_hash'] ?? null,
            ciRemote: $config['ci_remote'] ?? null,
            runtime: $config['runtime'],
            port: $config['port'] ?? null,
            image: $config['image'] ?? null,
            envVars: $config['env_vars'] ?? [],
            secrets: $config['secrets'] ?? [],
            outputDir: $config['output_dir'] ?? null,
            workingDir: $config['working_dir'] ?? null,
            runCmd: $config['run_cmd'] ?? null,
        );
    }

    public function processBlueprint()
    {
        $validation = $this->validateBlueprint();
        $spec = $this->getAttributes();
        $path = rtrim(self::BASE_PATH, '/').'/'.$spec->serviceName.'/'.ltrim($workingDir ?? '', '/');

        try {
            if (! $validation['valid']) {
                throw new \RuntimeException($validation['errors']);
            }

            $this->blueprint->updateOrFail(['status' => 'in_progress']);

            $remote = Remote::findOrFail($spec->remote->id);
            DirectoryService::setupFolder($path);

            // setup runtime

            $remoteService = new GitRepoService;
            $remoteService->cloneRepo($remote->name, $remote->repository, $remote->provider, $this->$path.'/'.$spec->serviceName);

            $newBlock = <<<EOF
            :$spec->port {
                root * $path/dist
                file_server
            }
            EOF;

            CaddyService::newConfig($spec->serviceName, $newBlock);

            if ($spec->runCmd !== null) {
                $cmd = CmdService::execute($spec->runCmd, ['working_directory' => $path]);
                $result = $cmd->successful;

                if (! $result) {
                    throw new \RuntimeException("Run command failed: {$spec->runCmd}");
                }
            }

            CaddyService::restart();

            $result = $this->blueprint->updateOrFail(['status' => 'completed']);

            if (! $result) {
                throw new \RuntimeException("Failed to update blueprint status for ID $spec->id", 1);
            }

            $service = Service::create(CleanParseService::withoutNulls(
                [
                    'name' => $spec->serviceName,
                    'source' => $remote->repository,
                    'runtime' => $spec->runtime,
                    'run_cmd' => $spec->runCmd,
                    'port' => $spec->port,
                    'working_dir' => $spec->workingDir,
                    'output_dir' => $spec->outputDir,
                    'image' => $spec->image,
                    'env_vars' => $spec->envVars,
                    'secrets' => $spec->secrets,
                    'remote_id' => $remote->id,
                    'ci_remote_id' => $spec->ciRemote->id,
                ]
            ));

            Log::info("Successfully created service $spec->serviceName. ID: ".$service->id);
        } catch (\RuntimeException $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            Log::error("Runtime exception on service $spec->serviceName: ".$e->getMessage());
        } catch (\Exception $e) {
            $this->blueprint->updateOrFail(['status' => 'failed']);
            $errorMessage = $e instanceof \Throwable ? $e->getMessage() : 'An unexpected error occurred.';
            Log::error("Failed to create service $spec->serviceName: ".$errorMessage);
        }
    }
}

<?php

namespace App\Http\Controllers\Projects\Services;

use App\Constants\Runtimes;
use App\Enums\JobStatus;
use App\Http\Controllers\Controller;
use App\Jobs\Services\ProcessBlueprint;
use App\Models\Blueprint;
use App\Models\Project;
use App\Models\Service;
use App\Services\CaddyService;
use App\Services\Secrets\SecretsManagerService;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Inertia\Inertia;

class ServicesController extends Controller
{
    /**
     * Show service creation page.
     */
    public function deploy(Project $project)
    {
        return Inertia::render('projects/services/deploy-service', [
            'project' => [
                'id' => $project->id,
                'name' => $project->name,
                'description' => $project->description,
            ],
        ]);
    }

    /* Fetch all services for a given project */
    public function fetch(Project $project): JsonResponse
    {
        return response()->json(
            Service::where('project_id', $project->id)
                ->get()
                ->map(fn ($service) => [
                    'id' => $service->id,
                    'name' => $service->name,
                    'source' => $service->source,
                    'runtime' => $service->runtime,
                    'run_cmd' => $service->run_cmd,
                    'build_cmd' => $service->build_cmd,
                    'port' => $service->port,
                    'working_dir' => $service->working_dir,
                    'static_dir' => $service->static_dir,
                    'image' => $service->image,
                    'spec' => $service->spec,
                    'env_vars' => $service->env_vars,
                    'secrets' => $service->secrets,
                    'remote_id' => $service->remote_id,
                    'ci_remote_id' => $service->ci_remote_id,
                ])
        );
    }

    /**
     * Show a service page.
     */
    public function show(Service $service)
    {
        return Inertia::render('projects/services/view-service', [
            'service' => [
                'id' => $service->id,
                'name' => $service->name,
                'source' => $service->source,
                'runtime' => $service->runtime,
                'run_cmd' => $service->run_cmd,
                'build_cmd' => $service->build_cmd,
                'port' => $service->port,
                'working_dir' => $service->working_dir,
                'static_dir' => $service->static_dir,
                'image' => $service->image,
                'spec' => $service->spec,
                'env_vars' => $service->env_vars,
                'secrets' => $service->secrets,
                'remote_id' => $service->remote_id,
                'ci_remote_id' => $service->ci_remote_id,
            ],
        ]);
    }

    public function checkPort(Request $request): RedirectResponse
    {
        $request->validate([
            'port' => ['required', 'integer', 'min:1', 'max:65535'],
        ]);

        $port = $request->input('port');

        $caddy = new CaddyService;

        if ($caddy->checkPort($port)) {
            return back()->withInput()->withErrors(['port' => __('Port '.$port.' is already in use. Choose another port.')]);
        }

        return back()->withInput();
    }

    /**
     * Handle a new service request.
     *
     * @throws \Illuminate\Validation\ValidationException
     */
    public function store(Request $request, Project $project): RedirectResponse
    {
        $request->validate([
            'name' => ['required', 'string'],
            'source' => ['required', 'in:remote,image'],
            'runtime' => ['required', 'in:static,go,php,python,node-js,ruby,dotnet,java,docker,k3s,custom'],
            'version' => ['nullable', 'string'],
            'run_cmd' => ['required_unless:runtime,static,docker,k3s', 'nullable', 'string'],
            'build_cmd' => ['nullable', 'string'],
            'port' => ['required_unless:runtime,static,docker,k3s', 'nullable', 'integer'],
            'working_dir' => ['nullable', 'string'],
            'static_dir' => ['nullable', 'string'],
            'image' => ['nullable', 'string'],
            'spec' => ['nullable', 'string'],
            'env_vars' => ['nullable', 'array'],
            'secrets' => ['nullable', 'array'],
            'remote' => ['nullable', 'string'],
            'ci_remote' => ['nullable', 'string'],
            'domain' => ['nullable', 'string'],
            'dns_provider' => ['nullable', 'string'],
        ]);

        $runtime = array_filter([
            'type' => $request->input('runtime'),
            'version' => $request->input('version'),
        ]);

        $config = array_filter([
            'name' => $request->input('name'),
            'source' => $request->input('source'),
            'runtime' => $runtime,
            'version' => $request->input('version'),
            'run_cmd' => $request->input('run_cmd'),
            'build_cmd' => $request->input('build_cmd'),
            'port' => $request->input('port'),
            'working_dir' => $request->input('working_dir'),
            'static_dir' => $request->input('static_dir'),
            'image' => $request->input('image'),
            'spec' => $request->input('spec'),
            'env_vars' => $request->input('env_vars'),
            'secrets' => $request->input('secrets'),
            'remote' => $request->input('remote'),
            'ci_remote' => $request->input('ci_remote'),
            'domain' => $request->input('domain'),
            'dns_provider' => $request->input('dns_provider'),
        ], fn ($value) => $value !== null);

        $metadata = array_filter([
            'project_id' => $project->id,
        ], fn ($value) => $value !== null);

        $blueprint = Blueprint::create([
            'status' => JobStatus::PENDING,
            'config' => $config,
            'metadata' => $metadata,
        ]);

        $name = $request->input('name');
        $envVars = $request->input('env_vars');
        $secrets = $request->input('secrets');

        if ($runtime['type'] !== Runtimes::STATIC && $runtime['type'] !== Runtimes::K3S && $runtime['type'] !== Runtimes::DOCKER) {
            $secretsManager = new SecretsManagerService;
            $secretsManager->tmp($envVars, $secrets, $name);
        }

        ProcessBlueprint::dispatch($blueprint);

        return back()->with('info', __('Creating service '.$request->name.' in progress.'));
    }
}

<?php

namespace App\Http\Controllers\Projects\Services;

use Inertia\Inertia;
use Illuminate\Http\Request;
use Illuminate\Http\RedirectResponse;
use App\Http\Controllers\Controller;
use App\Models\Service;
use App\Models\Project;
use App\Models\Blueprint;
use App\Enums\JobStatus;
use App\Jobs\Services\ProcessBlueprint;

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
                'port' => $service->port,
                'working_dir' => $service->working_dir,
                'output_dir' => $service->output_dir,
                'image' => $service->image,
                'spec' => $service->spec,
                'env_vars' => $service->env_vars,
                'secrets' => $service->secrets,
                'remote_id' => $service->remote_id,
                'ci_remote_id' => $service->ci_remote_id,
            ],
        ]);
    }

    /**
     * Handle a new service request.
     *
     * @throws \Illuminate\Validation\ValidationException
     */
    public function store(Request $request): RedirectResponse 
    {   
        $request->validate([
            'name' => ['required', 'string'],
            'source' => ['required', 'in:remote,image'],
            'runtime' => ['required', 'in:static,go,php,python,node-js,ruby,dotnet,java,docker,k3s,custom'],
            'run_cmd' => ['required_unless:runtime,static,docker,k3s', 'nullable', 'string'],
            'port' => ['required_unless:runtime,static,docker,k3s', 'nullable', 'integer'],
            'working_dir' => ['nullable', 'string'],
            'output_dir' => ['nullable', 'string'],
            'image' => ['nullable', 'string'],
            'spec' => ['nullable', 'string'],  
            'env_vars' => ['nullable', 'array'],
            'secrets' => ['nullable', 'array'],
            'remote' => ['nullable', 'string'],
            'ci_remote' => ['nullable', 'string'],
            'domain' => ['nullable', 'string'],
            'dns_provider' => ['nullable', 'string'],
        ]);

        $config = array_filter([
            'name' => $request->input('name'),
            'source' => $request->input('source'),
            'runtime' => $request->input('runtime'),
            'run_cmd' => $request->input('run_cmd'),
            'port' => $request->input('port'),
            'working_dir' => $request->input('working_dir'),
            'output_dir' => $request->input('output_dir'),
            'image' => $request->input('image'),
            'spec' => $request->input('spec'),
            'env_vars' => $request->input('env_vars'),
            'secrets' => $request->input('secrets'),
            'remote' => $request->input('remote'),
            'ci_remote' => $request->input('ci_remote'),
            'domain' => $request->input('domain'),
            'dns_provider' => $request->input('dns_provider'),
        ], fn($value) => $value !== null);

        Blueprint::create([
            'status' => JobStatus::PENDING,
            'config' => json_encode($config),
        ]);

        ProcessBlueprint::dispatch($config);

        return back()->with('info', __('Creating service '.$request->name.' in progress.'));
    }      
}
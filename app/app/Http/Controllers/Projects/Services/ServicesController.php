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
            'name' => ['required'],
            'source' => ['required', 'in:remote,image'],
            'runtime' => ['required', 'in:static,go,php,python,node-js,ruby,dotnet,java,docker,k3s,custom'],
            'run_cmd' => ['required_unless:runtime,static,docker,k3s'],
            'port' => ['required_unless:runtime,static,docker,k3s', 'integer'],
            'working_dir',
            'output_dir',
            'image',
            'spec',
            'env_vars',
            'secrets',
            'remote_id',
            'ci_remote_id',
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
            'remote_id' => $request->input('remote_id'),
            'ci_remote_id' => $request->input('ci_remote_id'),
        ], fn($value) => $value !== null);

        Blueprint::create([
            'status' => JobStatus::PENDING,
            'config' => json_encode($config),
        ]);

        return back()->with('info', __('Creating service '.$request->name.' in progress.'));
    }      
}
<?php

namespace Tests\Feature;

use App\Enums\JobStatus;
use App\Jobs\Services\ProcessBlueprint;
use App\Models\Blueprint;
use App\Models\Project;
use App\Models\Service;
use App\Models\User;
use App\Services\CaddyService;
use Illuminate\Foundation\Testing\DatabaseMigrations;
use Illuminate\Support\Facades\Queue;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class ServicesTest extends TestCase
{
    use DatabaseMigrations;

    // Ensure all static tests are cleaned-up after
    public function tearDown(): void
    {
        \Mockery::close();
        parent::tearDown();
    }

    public function test_authenticated_users_can_view_service_deploy_page()
    {
        $this->actingAs(User::factory()->create());
        $project = Project::factory()->create();

        $this->get(route('servicesIndex', $project))
            ->assertInertia(fn (Assert $page) =>
                $page->component('projects/services/deploy-service')
                    ->has('project', fn ($p) =>
                        $p->where('id', $project->id)
                          ->where('name', $project->name)
                          ->etc()
                    )
            )
            ->assertOk();
    }

    public function test_it_fetches_all_services_for_a_project_as_json()
    {
        $this->actingAs(User::factory()->create());

        $project = Project::factory()->create();
        Service::factory()->count(2)->create(['project_id' => $project->id]);

        $this->getJson(route('servicesList', $project))
            ->assertOk()
            ->assertJsonStructure([
                '*' => [
                    'id', 'name', 'source', 'runtime', 'run_cmd', 'port', 
                ]
            ])
            ->assertJsonCount(2);
    }

    public function test_it_validates_port_field_when_checking_port_availability()
    {
        $this->actingAs(User::factory()->create());

        $this->post(route('servicesCheckPort'), [])
            ->assertSessionHasErrors('port');

        $this->post(route('servicesCheckPort'), ['port' => 70000])
            ->assertSessionHasErrors('port');
    }

    public function test_it_returns_error_if_port_is_in_use()
    {
        $this->actingAs(User::factory()->create());
        \Mockery::mock('alias:' . CaddyService::class)
            ->shouldReceive('checkPort')
            ->once()
            ->andReturn(true);

        $this->post(route('servicesCheckPort'), ['port' => 8080])
            ->assertSessionHasErrors('port');
    }

    public function test_it_passes_validation_and_returns_back_if_port_is_free()
    {
        $this->actingAs(User::factory()->create());

        \Mockery::mock('alias:' . CaddyService::class)
            ->shouldReceive('checkPort')
            ->once()
            ->andReturn(false);

        $this->post(route('servicesCheckPort'), ['port' => 8080])
            ->assertRedirect();
    }

    public function test_it_stores_a_service_request_and_dispatches_a_blueprint_job()
    {
        $this->actingAs(User::factory()->create());
        $project = Project::factory()->create();

        Queue::fake();

        $payload = [
            'name' => 'My App',
            'source' => 'remote',
            'runtime' => 'node-js',
            'run_cmd' => 'npm start',
            'port' => 3000,
            'working_dir' => '/app',
        ];

        $this->post(route('servicesCreate', ['project' => $project->id]), $payload)
            ->assertRedirect()
            ->assertSessionHas('info', 'Creating service My App in progress.');

        $this->assertDatabaseHas('blueprints', [
            'status' => JobStatus::PENDING,
        ]);

        Queue::assertPushed(ProcessBlueprint::class);
    }

    public function test_form_validation()
    {
        $this->actingAs(User::factory()->create());
        $project = Project::factory()->create();

        $this->post(route('servicesCreate', ['project' => $project->id]), [])
            ->assertSessionHasErrors(['name', 'source', 'runtime']);
    }
}

<?php

namespace Tests\Feature;

use App\Models\Blueprint;
use App\Models\User;
use Illuminate\Foundation\Testing\DatabaseMigrations;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class BlueprintsTest extends TestCase
{
    use DatabaseMigrations;

    public function test_authenticated_users_can_view_the_deployments_page()
    {
        $this->actingAs(User::factory()->create());

        $this->get(route('deploymentsIndex'))
            ->assertInertia(fn (Assert $page) => $page->component('deployments/index')
            )
            ->assertOk();
    }

    public function test_it_fetches_all_blueprints_as_json()
    {
        $this->actingAs(User::factory()->create());

        Blueprint::factory()->count(3)->create();

        $response = $this->getJson(route('deploymentsList'));

        $response
            ->assertOk()
            ->assertJsonStructure([
                '*' => ['id', 'config', 'status', 'created_at', 'updated_at'],
            ])
            ->assertJsonCount(3);
    }

    public function test_authenticated_users_can_view_a_single_blueprint_page()
    {
        $this->actingAs(User::factory()->create());

        $blueprint = Blueprint::factory()->create();

        $this->get(route('deploymentsShow', $blueprint))
            ->assertInertia(fn (Assert $page) => $page->component('deployments/view-deployment')
                ->has('blueprint', fn ($bp) => $bp
                    ->where('id', $blueprint->id)
                    ->where('status', 'pending')
                    ->etc()
                )
            )
            ->assertOk();
    }
}

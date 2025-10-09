<?php

namespace Tests\Feature;

use App\Models\Project;
use App\Models\User;
use Illuminate\Foundation\Testing\DatabaseMigrations;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class DashboardTest extends TestCase
{
    use DatabaseMigrations;

    public function test_it_fetches_all_projects_as_json()
    {
        $this->actingAs(User::factory()->create());

        $this->get(route('projectsList'))->assertJsonStructure([
            '*' => ['id', 'name', 'description'],
        ]);
    }

    public function test_authenticated_users_can_view_a_project_page()
    {
        $this->actingAs(User::factory()->create());

        $project = Project::factory()->create();

        $this->get(route('projectsShow', $project))
            ->assertInertia(fn (Assert $page) => $page->component('projects/services/index')
                ->has('project', fn (Assert $p) => $p->where('id', $project->id)
                    ->where('name', $project->name)
                    ->where('description', $project->description)
                    ->etc()
                )
            );
    }
}

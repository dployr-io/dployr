<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\DatabaseMigrations;
use Tests\TestCase;

class ProjectsTest extends TestCase
{
    use DatabaseMigrations;

    public function test_guests_are_redirected_to_the_login_page()
    {
        $this->get(route('projectsIndex'))->assertRedirect(route('login'));
    }

    public function test_authenticated_users_can_visit_the_dashboard()
    {
        $this->actingAs(User::factory()->create());

        $this->get(route('projectsIndex'))->assertOk();
    }

    public function test_authenticated_users_can_create_a_project()
    {
        $this->actingAs(User::factory()->create());

        $response = $this->post(route('projectsCreate'), [
            'name' => 'My awesome dployr app',
            'description' => 'Your app, your server, your rules!',
        ]);

        $response
            ->assertRedirect()
            ->assertSessionHas('success', 'Your project was created successfully.');

        $this->assertDatabaseHas('projects', [
            'name' => 'My awesome dployr app',
            'description' => 'Your app, your server, your rules!',
        ]);
    }
}

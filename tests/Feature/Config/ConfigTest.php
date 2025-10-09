<?php

namespace Tests\Feature;

use App\Models\Config;
use App\Models\User;
use Illuminate\Foundation\Testing\DatabaseMigrations;
use Tests\TestCase;

class ConfigTest extends TestCase
{
    use DatabaseMigrations;

    public function test_it_stores_provided_tokens_and_returns_success()
    {
        $this->actingAs(User::factory()->create());

        $payload = [
            'github_token' => 'ghp_123456',
            'gitlab_token' => null,
            'bitbucket_token' => null,
            'slack_api_key' => null,
        ];

        $response = $this->post(route('updateConfig'), $payload);

        $response->assertRedirect();
        $response->assertSessionHas('success', 'Configuration updated successfully!');

        $config = Config::where('key', 'github_token')->first();

        $this->assertEquals('ghp_123456', $config->value);
    }

    public function test_it_updates_existing_config_values()
    {
        Config::create(['key' => 'slack_api_key', 'value' => 'old_key']);

        $this->actingAs(User::factory()->create())
            ->post(route('updateConfig'), [
                'slack_api_key' => 'new_key',
            ])
            ->assertRedirect()
            ->assertSessionHas('success');

        $config = Config::where('key', 'slack_api_key')->first();

        $this->assertEquals('new_key', $config->value);
    }
}

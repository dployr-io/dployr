<?php

namespace Tests\Feature;

use App\Facades\AppConfig;
use App\Models\Remote;
use App\Models\User;
use App\Services\GitRepoService;
use App\Services\HttpService;
use Illuminate\Foundation\Testing\DatabaseMigrations;
use Inertia\Testing\AssertableInertia as Assert;
use Tests\TestCase;

class RemotesTest extends TestCase
{
    use DatabaseMigrations;

    // Ensure all static tests are cleaned-up after
    protected function tearDown(): void
    {
        if ($this->app) {
            \DB::rollBack();
        }
        \Mockery::close();
        parent::tearDown();
    }

    public function test_it_renders_the_remotes_index_page()
    {
        $this->actingAs(User::factory()->create())
            ->get(route('remotesIndex'))
            ->assertInertia(fn (Assert $page) => $page->component('resources/remotes')
            )
            ->assertOk();
    }

    public function test_it_returns_paginated_remotes_as_json()
    {
        $this->actingAs(User::factory()->create());
        Remote::factory()->count(25)->create();

        $this->mock(GitRepoService::class, fn ($mock) => $mock->shouldReceive('getLatestCommitMessage')
            ->andReturn([
                'message' => 'Test commit message',
                'avatar_url' => 'https://example.com/avatar.png',
            ])
        );

        $response = $this->getJson(route('remotesFetch'));

        $response->assertOk()
            ->assertJsonStructure([
                '*' => ['id', 'name', 'repository', 'provider', 'avatar_url'],
            ]);
    }

    public function test_it_handles_commit_fetch_failure_gracefully()
    {
        $this->actingAs(User::factory()->create());
        Remote::factory()->create();

        $this->mock(GitRepoService::class, fn ($mock) => $mock->shouldReceive('getLatestCommitMessage')
            ->andThrow(new \RuntimeException('API Down'))
        );

        $response = $this->getJson(route('remotesFetch'));
        $response->assertOk();
        $this->assertEquals('Unable to fetch commit', $response->json()[0]['commit_message']);
    }

    // public function test_it_renders_a_single_remote_page_with_commit_data()
    // {
    //     $this->actingAs(User::factory()->create());
    //     $remote = Remote::factory()->create();

    //     \Mockery::mock('alias:' . HttpService::class)
    //         ->shouldReceive('makeRequest')
    //         ->andThrow(new \RuntimeException('Repo not found'));

    //     $response = $this->post(route('remotesShow', $remote))
    //         ->assertSessionHas('remote');

    //     $response->assertOk();
    // }

    public function test_it_validates_required_fields_on_store()
    {
        $this->actingAs(User::factory()->create());

        $this->post(route('remotesCreate'), [])
            ->assertSessionHasErrors(['remote_repo', 'branch']);
    }

    public function test_it_creates_a_new_remote_successfully()
    {
        $this->actingAs(User::factory()->create());

        AppConfig::shouldReceive('get')
            ->with('github_token')
            ->andReturn('fake-token');

        \Mockery::mock('alias:'.HttpService::class)
            ->shouldReceive('makeRequest')
            ->andReturn([
                'name' => 'repo-name',
                'repository' => 'org/repo',
                'provider' => 'github',
            ]);

        $response = $this->post(route('remotesCreate'), [
            'remote_repo' => 'https://github.com/org/repo',
            'branch' => 'main',
        ]);

        $response->assertRedirectBack();
    }

    public function test_it_handles_search_with_validation_and_service_response()
    {
        $this->actingAs(User::factory()->create());

        \Mockery::mock('alias:'.HttpService::class)
            ->shouldReceive('makeRequest')
            ->andReturnUsing(function ($method, $url, $headers, $type) {
                if ($type === 'repository') {
                    return [
                        'owner' => ['avatar_url' => 'https://avatar.url'],
                        'clone_url' => 'https://github.com/org/repo.git',
                    ];
                } elseif ($type === 'branches') {
                    return [
                        ['name' => 'main'],
                        ['name' => 'develop'],
                    ];
                } elseif ($type === 'commits') {
                    return [
                        ['commit' => ['message' => 'Initial commit'], 'author' => ['avatar_url' => 'https://avatar.url']],
                    ];
                }

                return [];
            });

        $response = $this->post(route('remotesSearch'), [
            'remote_repo' => 'github.com/org/repo',
        ]);

        $response->assertRedirectBack();
    }

    public function test_it_handles_search_service_failure()
    {
        $this->actingAs(User::factory()->create());

        AppConfig::shouldReceive('get')
            ->once()
            ->with('github_token')
            ->andReturn('fake-token');

        \Mockery::mock('alias:'.HttpService::class)
            ->shouldReceive('makeRequest')
            ->andThrow(new \RuntimeException('Repo not found'));

        $this->post(route('remotesSearch'), [
            'remote_repo' => 'github.com/org/repo',
        ])->assertSessionHas('error', 'Repo not found');
    }
}

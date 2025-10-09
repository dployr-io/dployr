<?php

namespace Database\Factories;

use Illuminate\Database\Eloquent\Factories\Factory;
use Symfony\Component\Uid\Ulid;

class ServiceFactory extends Factory
{
    /**
     * Define the model's default state.
     *
     * @return array<string, mixed>
     */
    public function definition(): array
    {
        return [
            'id' => Ulid::generate(),
            'name' => fake()->name(),
            'source' => 'remote',
            'runtime' => 'node-js',
            'run_cmd' => 'npm run start',
            'port' => 3000,
        ];
    }
}
